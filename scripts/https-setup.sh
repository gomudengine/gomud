#!/usr/bin/env sh

set -eu

CONFIG_FILE="${CONFIG_FILE:-_datafiles/config.yaml}"
CURL_BIN="${CURL_BIN:-curl}"

if [ ! -f "$CONFIG_FILE" ]; then
	echo "Config file not found: $CONFIG_FILE" >&2
	exit 1
fi

read_yaml_value() {
	read_yaml_value_from_file "$CONFIG_FILE" "$1"
}

read_yaml_value_from_file() {
	file_path=$1
	key=$2
	if [ ! -f "$file_path" ]; then
		return 0
	fi
	awk -F': ' -v key="$key" '
		$1 ~ "^[[:space:]]*" key "$" {
			gsub(/^["'\'']|["'\'']$/, "", $2)
			print $2
			exit
		}
	' "$file_path"
}

canonicalize_path() {
	path=$1
	path_dir=$(dirname "$path")
	path_base=$(basename "$path")
	if abs_dir=$(cd "$path_dir" 2>/dev/null && pwd -P); then
		printf '%s/%s\n' "$abs_dir" "$path_base"
		return 0
	fi
	printf '%s\n' "$path"
}

json_escape() {
	printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

yaml_single_quote() {
	printf "'%s'" "$(printf '%s' "$1" | sed "s/'/''/g")"
}

base_url_without_trailing_slash() {
	url=$1
	while [ "${url%/}" != "$url" ]; do
		url=${url%/}
	done
	printf '%s' "$url"
}

current_data_files=$(read_yaml_value DataFiles)
default_override_file="${current_data_files:-_datafiles/world/default}/config-overrides.yaml"
override_file="${CONFIG_PATH:-$default_override_file}"
bundled_config_path=$(canonicalize_path "$CONFIG_FILE")
override_path=$(canonicalize_path "$override_file")

if [ "$override_path" = "$bundled_config_path" ]; then
	printf 'Ignoring CONFIG_PATH=%s because https-setup must not target the bundled base config.\n' "$override_file" >&2
	override_file=$default_override_file
fi

read_effective_yaml_value() {
	key=$1
	value=$(read_yaml_value_from_file "$override_file" "$key")
	if [ -n "$value" ]; then
		printf '%s\n' "$value"
		return 0
	fi
	read_yaml_value "$key"
}

current_https_cert=$(read_effective_yaml_value HttpsCertFile)
current_https_key=$(read_effective_yaml_value HttpsKeyFile)
current_http_port=$(read_effective_yaml_value HttpPort)
current_https_port=$(read_effective_yaml_value HttpsPort)
current_https_redirect=$(read_effective_yaml_value HttpsRedirect)

https_cert_file=$current_https_cert
https_key_file=$current_https_key
http_port=$current_http_port
https_port=$current_https_port
https_redirect=$current_https_redirect
config_updates=""
override_snippet=""

printf 'Interactive HTTPS setup\n'
printf 'Bundled base config: %s\n' "$CONFIG_FILE"
printf 'Override target: %s\n\n' "$override_file"

printf 'Choose HTTPS mode:\n'
printf '  1) Manual certificate files\n'
printf '  2) HTTP only\n'
printf 'Selection [1]: '
IFS= read -r mode_selection
mode_selection=${mode_selection:-1}

case "$mode_selection" in
1)
	printf 'Certificate file path [%s]: ' "${current_https_cert:-server.crt}"
	IFS= read -r cert_input
	if [ -n "$cert_input" ]; then
		https_cert_file=$cert_input
	elif [ -z "$https_cert_file" ]; then
		https_cert_file=server.crt
	fi

	printf 'Private key file path [%s]: ' "${current_https_key:-server.key}"
	IFS= read -r key_input
	if [ -n "$key_input" ]; then
		https_key_file=$key_input
	elif [ -z "$https_key_file" ]; then
		https_key_file=server.key
	fi

	printf 'HTTP port [%s]: ' "${current_http_port:-80}"
	IFS= read -r http_port_input
	if [ -n "$http_port_input" ]; then
		http_port=$http_port_input
	elif [ -z "$http_port" ]; then
		http_port=80
	fi

	printf 'HTTPS port [%s]: ' "${current_https_port:-443}"
	IFS= read -r https_port_input
	if [ -n "$https_port_input" ]; then
		https_port=$https_port_input
	elif [ -z "$https_port" ]; then
		https_port=443
	fi

	printf 'Redirect HTTP to HTTPS? [%s]: ' "${current_https_redirect:-true}"
	IFS= read -r redirect_input
	if [ -n "$redirect_input" ]; then
		https_redirect=$redirect_input
	else
		https_redirect=${current_https_redirect:-true}
	fi
	;;
2)
	printf 'HTTP port [%s]: ' "${current_http_port:-80}"
	IFS= read -r http_port_input
	if [ -n "$http_port_input" ]; then
		http_port=$http_port_input
	elif [ -z "$http_port" ]; then
		http_port=80
	fi

	https_port=0
	https_redirect=false
	https_cert_file=""
	https_key_file=""
	;;
*)
	echo "Unknown selection: $mode_selection" >&2
	exit 1
	;;
esac

config_updates=$(
	cat <<EOF
"FilePaths.HttpsCertFile":"$(json_escape "$https_cert_file")"
,"FilePaths.HttpsKeyFile":"$(json_escape "$https_key_file")"
,"Network.HttpPort":"$(json_escape "$http_port")"
,"Network.HttpsPort":"$(json_escape "$https_port")"
,"Network.HttpsRedirect":"$(json_escape "$https_redirect")"
EOF
)

override_snippet=$(
	cat <<EOF
FilePaths:
  HttpsCertFile: $(yaml_single_quote "$https_cert_file")
  HttpsKeyFile: $(yaml_single_quote "$https_key_file")
Network:
  HttpPort: $http_port
  HttpsPort: $https_port
  HttpsRedirect: $https_redirect
EOF
)

printf '\nPlanned settings:\n'
printf '  HttpsCertFile: %s\n' "${https_cert_file:-<empty>}"
printf '  HttpsKeyFile: %s\n' "${https_key_file:-<empty>}"
printf '  HttpPort: %s\n' "$http_port"
printf '  HttpsPort: %s\n' "$https_port"
printf '  HttpsRedirect: %s\n' "$https_redirect"

if [ "$mode_selection" = "1" ]; then
	printf '\nBefore applying these settings:\n'
	printf '  - Make sure %s and %s exist and are readable by the server.\n' "$https_cert_file" "$https_key_file"
	printf '  - Open inbound TCP port %s if players should use HTTPS from the internet.\n' "$https_port"
fi

printf '\nChoose how to apply these changes:\n'
printf '  1) PATCH a running GoMud server via /admin/api/v1/config\n'
printf '  2) Print a config-overrides snippet for manual save\n'
printf 'Selection [2]: '
IFS= read -r apply_selection
apply_selection=${apply_selection:-2}

case "$apply_selection" in
1)
	printf 'Admin base URL [http://localhost]: '
	IFS= read -r admin_base_url
	admin_base_url=${admin_base_url:-http://localhost}
	admin_base_url=$(base_url_without_trailing_slash "$admin_base_url")

	printf 'Admin username [admin]: '
	IFS= read -r admin_username
	admin_username=${admin_username:-admin}

	printf 'Admin password: '
	IFS= read -r admin_password

	printf '\nPATCH %s/admin/api/v1/config\n' "$admin_base_url"
	if ! "$CURL_BIN" --fail --silent --show-error \
		-u "$admin_username:$admin_password" \
		-H 'Content-Type: application/json' \
		-X PATCH \
		--data "{$config_updates}" \
		"$admin_base_url/admin/api/v1/config"; then
		printf '\nFailed to apply settings through the admin API.\n' >&2
		printf 'Fallback: save the following override snippet to %s and restart GoMud:\n\n' "$override_file" >&2
		printf '%s\n' "$override_snippet" >&2
		exit 1
	fi

	printf '\nHTTPS setup applied through the admin API.\n'
	printf 'Next steps:\n'
	if [ "$mode_selection" = "1" ]; then
		printf '  1. Confirm %s and %s exist and are readable.\n' "$https_cert_file" "$https_key_file"
		printf '  2. Restart GoMud only if your deployment requires it, then open https://your-domain:%s/.\n' "$https_port"
	else
		printf '  1. Restart GoMud only if your deployment requires it, then connect over plain HTTP on port %s.\n' "$http_port"
	fi
	;;
2)
	printf '\nSave the following override snippet to %s:\n\n' "$override_file"
	printf '%s\n' "$override_snippet"
	printf '\nNext steps:\n'
	if [ "$mode_selection" = "1" ]; then
		printf '  1. Save the snippet to %s or set the same keys through the admin API.\n' "$override_file"
		printf '  2. Confirm %s and %s exist and are readable.\n' "$https_cert_file" "$https_key_file"
		printf '  3. Restart GoMud and open https://your-domain:%s/.\n' "$https_port"
	else
		printf '  1. Save the snippet to %s or set the same keys through the admin API.\n' "$override_file"
		printf '  2. Restart GoMud and connect over plain HTTP on port %s.\n' "$http_port"
	fi
	;;
*)
	echo "Unknown selection: $apply_selection" >&2
	exit 1
	;;
esac
