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
