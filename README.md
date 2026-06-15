# gosshd

[English](README.md) | [简体中文](README.zh-CN.md)

`gosshd` is a minimal Go SSH relay: a public server accepts standard SSH/SFTP/SCP/tunnel clients, while a private-network agent opens one outbound WebSocket connection and exposes the local machine through a stable UUID.

The UUID is the v1 access secret. Anyone who knows it can access the agent machine with the permissions of the running `gosshd-agent` process.

## Architecture

```text
                         standard SSH / SFTP / SCP / tunnels
  +------------------+   ssh UUID@public-host    +----------------------+
  | SSH client       | -------------------------> | gosshd-server        |
  | anywhere         |                            | public network       |
  +------------------+                            | HTTP :80 / SSH :22   |
                                                  +----------+-----------+
                                                             ^
                                                             |
                                      outbound WebSocket     |
                                      + yamux streams        |
                                                             |
                                                  +----------+-----------+
                                                  | gosshd-agent         |
                                                  | private network      |
                                                  | shell / SFTP / TCP   |
                                                  +----------+-----------+
                                                             |
                                                  +----------v-----------+
                                                  | private host         |
                                                  +----------------------+
```

## Quick Start From GitHub Releases

This path uses prebuilt GitHub Release archives for the public server. No local Go toolchain or local build is required.

Pick the server archive that matches the public server from the [latest release](https://github.com/qinyongliang/gosshd/releases/latest). The examples below use `linux-amd64` and `v0.1.1`; replace them when using a different release or platform.

Start the public server:

```sh
VERSION=v0.1.1
PLATFORM=linux-amd64
curl -fSLO "https://github.com/qinyongliang/gosshd/releases/download/${VERSION}/gosshd-${VERSION}-${PLATFORM}.tar.gz"
tar -xzf "gosshd-${VERSION}-${PLATFORM}.tar.gz"
cd "gosshd-${PLATFORM}"
sudo ./gosshd-server --http-listen :80 --ssh-listen :22 --public-host public-host
```

The release server archive includes the agent download tree under `dist/agent`, so private-network machines can install the correct agent directly from your running `gosshd-server`.

Start an agent on a private-network Linux/macOS host:

```sh
curl http://public-host/install.sh | sh
```

Start an agent on a private-network Windows host:

```powershell
irm http://public-host/install.ps1 | iex
```

The agent prints an address like:

```text
ssh UUID@public-host
```

Then connect from anywhere:

```sh
ssh UUID@public-host
sftp UUID@public-host
scp file UUID@public-host:/tmp/file
ssh -L 15432:127.0.0.1:5432 UUID@public-host
ssh -D 1080 UUID@public-host
ssh -R 0:127.0.0.1:8080 UUID@public-host
```

If the public server maps gosshd SSH to a non-default port, add `-p`/`-P`, for example `ssh -p 2222 UUID@public-host`.

## Build

```powershell
go mod tidy
go build -o bin/gosshd-server.exe ./cmd/gosshd-server
go build -o bin/gosshd-agent.exe ./cmd/gosshd-agent
```

Cross-build agent binaries for server downloads:

```powershell
$env:GOOS='linux'; $env:GOARCH='amd64'; go build -o dist/agent/linux/amd64/gosshd-agent ./cmd/gosshd-agent
$env:GOOS='windows'; $env:GOARCH='amd64'; go build -o dist/agent/windows/amd64/gosshd-agent.exe ./cmd/gosshd-agent
Remove-Item Env:\GOOS,Env:\GOARCH
```

Release archives are built by GitHub Actions when a GitHub Release is created, covering common Linux, Windows, macOS, FreeBSD, OpenBSD, and NetBSD CPU architectures. Each server archive includes `dist/agent` binaries used by `/install.sh`, `/install.ps1`, and `/download/agent/{goos}/{goarch}`.

## Run

Development ports:

```powershell
bin/gosshd-server.exe --http-listen :8080 --ssh-listen :2222 --public-host localhost:8080 --agent-path dist/agent
bin/gosshd-agent.exe --server http://localhost:8080
```

Production defaults are configurable and default to `:80` for HTTP and `:22` for SSH.

## Docker Server

Build a Linux server image with downloadable agents for the supported OS/CPU matrix:

```powershell
docker build -t gosshd-server:latest .
```

Run locally on high ports:

```powershell
docker run --rm -p 8080:80 -p 2222:22 gosshd-server:latest --public-host localhost:8080 --http-listen :80 --ssh-listen :22 --agent-path /app/agent
```

Run on a public host with default ports:

```sh
docker run -d --name gosshd-server --restart unless-stopped \
  -p 80:80 -p 22:22 \
  gosshd-server:latest \
  --public-host your.host.name --http-listen :80 --ssh-listen :22 --agent-path /app/agent
```

If host SSH already uses port `22`, map the relay SSH port to a high port first, for example `-p 2222:22`.

Linux/macOS bootstrap:

```sh
curl http://public-host/install.sh | sh
```

Windows bootstrap:

```powershell
irm http://public-host/install.ps1 | iex
```

The agent prints an address like:

```text
ssh UUID@public-host
```

When using non-default SSH ports:

```sh
ssh -p 2222 UUID@public-host
sftp -P 2222 UUID@public-host
scp -P 2222 file UUID@public-host:/tmp/file
ssh -p 2222 -L 15432:127.0.0.1:5432 UUID@public-host
ssh -p 2222 -D 1080 UUID@public-host
ssh -p 2222 -R 0:127.0.0.1:8080 UUID@public-host
```

## Release

Create a GitHub Release and `.github/workflows/release.yml` builds release archives for the supported OS/CPU matrix and uploads them as release assets.

The workflow can also be run manually from the GitHub Actions page for a packaging smoke test.

## v1 Notes

- Agents are temporary foreground processes, not system services.
- Agent UUIDs are stored under `~/.gosshd/agent.json` by default.
- Server state is in-memory; offline agents are forgotten.
- SFTP exposes the filesystem visible to the agent process.
- Remote forwarding only binds `127.0.0.1`/`localhost` on the public server.
- TLS, Web UI, audit logs, multi-user auth, and permanent service installation are out of scope for v1.
