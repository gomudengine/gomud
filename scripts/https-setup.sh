#!/usr/bin/env sh

set -eu

CONFIG_FILE="${CONFIG_FILE:-_datafiles/config.yaml}"

if [ ! -f "$CONFIG_FILE" ]; then
	echo "Config file not found: $CONFIG_FILE" >&2
	exit 1
fi

read_yaml_value() {
	awk -F': ' -v key="$1" '
		$1 ~ "^[[:space:]]*" key "$" {
			gsub(/^"|"$/, "", $2)
			print $2
			exit
		}
	' "$CONFIG_FILE"
}

current_https_cert=$(read_yaml_value HttpsCertFile)
current_https_key=$(read_yaml_value HttpsKeyFile)
current_http_port=$(read_yaml_value HttpPort)
current_https_port=$(read_yaml_value HttpsPort)
current_https_redirect=$(read_yaml_value HttpsRedirect)

https_cert_file=$current_https_cert
https_key_file=$current_https_key
http_port=$current_http_port
https_port=$current_https_port
https_redirect=$current_https_redirect

printf 'Interactive HTTPS setup\n'
printf 'Config file: %s\n\n' "$CONFIG_FILE"

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

printf '\nPlanned settings:\n'
printf '  HttpsCertFile: %s\n' "${https_cert_file:-<empty>}"
printf '  HttpsKeyFile: %s\n' "${https_key_file:-<empty>}"
printf '  HttpPort: %s\n' "$http_port"
printf '  HttpsPort: %s\n' "$https_port"
printf '  HttpsRedirect: %s\n' "$https_redirect"

if [ "$mode_selection" = "1" ]; then
	printf '\nBefore restarting GoMud:\n'
	printf '  - Make sure %s and %s exist and are readable by the server.\n' "$https_cert_file" "$https_key_file"
	printf '  - Open inbound TCP port %s if players should use HTTPS from the internet.\n' "$https_port"
fi

printf '\nApply these changes to %s? [Y/n]: ' "$CONFIG_FILE"
IFS= read -r confirm
confirm=${confirm:-Y}
case "$confirm" in
Y | y | yes | YES)
	;;
*)
	echo "Aborted without changing the config."
	exit 0
	;;
esac

timestamp=$(date +%Y%m%d%H%M%S)
backup_file="${CONFIG_FILE}.bak.${timestamp}"
cp "$CONFIG_FILE" "$backup_file"

tmp_file=$(mktemp "${TMPDIR:-/tmp}/gomud-https-setup.XXXXXX")
cleanup() {
	if [ -n "${tmp_file:-}" ]; then
		rm -f "$tmp_file"
	fi
}
trap cleanup EXIT HUP INT TERM

awk \
	-v https_cert_file="$https_cert_file" \
	-v https_key_file="$https_key_file" \
	-v http_port="$http_port" \
	-v https_port="$https_port" \
	-v https_redirect="$https_redirect" \
	'
	function yaml_string(s, escaped) {
		escaped = s
		gsub(/\\/,"\\\\", escaped)
		gsub(/"/,"\\\"", escaped)
		return "\"" escaped "\""
	}
	/^[[:space:]]*HttpsCertFile:/ { print "  HttpsCertFile: " yaml_string(https_cert_file); next }
	/^[[:space:]]*HttpsKeyFile:/  { print "  HttpsKeyFile: " yaml_string(https_key_file); next }
	/^[[:space:]]*HttpPort:/      { print "  HttpPort: " http_port; next }
	/^[[:space:]]*HttpsPort:/     { print "  HttpsPort: " https_port; next }
	/^[[:space:]]*HttpsRedirect:/ { print "  HttpsRedirect: " https_redirect; next }
	{ print }
	' "$CONFIG_FILE" >"$tmp_file"

# Rewrite the existing file in place so we preserve its owner and mode.
cat "$tmp_file" >"$CONFIG_FILE"

printf '\nHTTPS setup updated.\n'
printf 'Backup saved to: %s\n' "$backup_file"
printf 'Next steps:\n'
case "$mode_selection" in
1)
	printf '  1. Confirm %s and %s exist and are readable.\n' "$https_cert_file" "$https_key_file"
	printf '  2. Restart GoMud and open https://your-domain:%s/.\n' "$https_port"
	;;
2)
	printf '  1. Restart GoMud and connect over plain HTTP on port %s.\n' "$http_port"
	;;
esac
