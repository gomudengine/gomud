# Getting Running on Various Platforms

These guides assume you want to build from the source. You can also download a
pre-compiled package from the [releases section](https://github.com/GoMudEngine/GoMud/releases)
of the repo. Use the rolling `prerelease` entry for the latest `master` build,
or choose a numbered release for a permanent versioned build.

- [Raspberry PI Zero 2W](RASPBERRY-PI.md)
- [Running via Docker](DOCKER.md)
- [Setting Up an EC2 Instance](EC2.md)


# Quick Start

You can download a build from the [releases page](https://github.com/GoMudEngine/GoMud/releases),
unzip it and run the binary to get started, or if you prefer to build it
yourself, follow the instructions below. Use `prerelease` for the newest
`master` build, or a numbered release if you want a stable version you can
return to later.

A youtube playlist to getting started has been set up here:

[![Getting Started Videos](https://i.ytimg.com/vi/OOZqX01aHt8/hqdefault.jpg "Getting Started Playlist")](https://www.youtube.com/watch?v=OOZqX01aHt8&list=PL20JEmG_bxBuaOE9oFziAhAmx1pyXhQ1p)

You can compile and run it locally with:

> `go run .`

Or you can just build the binary if you prefer:

> `go build -o GoMudServer`

> `./GoMudServer`

Or if you have docker installed:

> `docker compose up --build`


# HTTPS With Certificate Files

GoMud already supports HTTPS when you provide certificate files.

For a guided config update, run:

> `make https-setup`

The helper does not edit the bundled base config directly.
It can PATCH a running GoMud server through `/admin/api/v1/config`, or print a `config-overrides.yaml` snippet for manual save.
It can configure manual certificate files, automatic Let's Encrypt, or disable HTTPS and return to HTTP-only mode.
Either path still requires a GoMud restart before listener changes take effect.

1. Get a certificate and private key for the hostname players will use.
2. Set `FilePaths.HttpsCertFile` to the certificate path.
3. Set `FilePaths.HttpsKeyFile` to the private key path.
4. Set `Network.HttpsPort` to the HTTPS port.
5. Optionally set `Network.HttpsRedirect` to `true`.

Restart GoMud after applying the settings.

# Automatic HTTPS

For simple single-server installs, GoMud can automatically manage Let's Encrypt certificates for the built-in web client and admin UI.

1. Point a public DNS name at your server.
2. Run `make https-setup` and choose `Automatic Let's Encrypt`, then either PATCH the running server or save the printed override snippet.
3. Set `FilePaths.WebDomain` to that hostname.
4. Set `Network.HttpPort` to `80` and `Network.HttpsPort` to `443`.
5. Optionally set `FilePaths.HttpsEmail` so Let's Encrypt can send expiry notices.
6. Leave `FilePaths.HttpsCertFile` and `FilePaths.HttpsKeyFile` empty unless you want to use your own certificate files instead.

If HTTPS is not working and you need to roll back quickly, run `make https-setup`, choose `Disable HTTPS and use HTTP only`, apply the change, and restart GoMud so it rebinds the listeners.

Notes:

- Automatic HTTPS is intended for one public server that owns ports `80` and `443`.
- `localhost`, private-only names, and raw IP addresses will stay on HTTP.
- If automatic HTTPS cannot succeed, GoMud falls back to HTTP and logs what to fix.
- The certificate cache is stored in `FilePaths.HttpsCacheDir`, which defaults to `_datafiles/tls`.
- The admin page at `/admin/https/` shows the current HTTPS mode, checks, and recommended fixes.
