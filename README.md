## Discord-bot

[![GitHub release](https://img.shields.io/github/release/blueprintue/discord-bot.svg?logo=github)](https://github.com/blueprintue/discord-bot/releases/latest)
[![Total downloads](https://img.shields.io/github/downloads/blueprintue/discord-bot/total.svg?logo=github)](https://github.com/blueprintue/discord-bot/releases/latest)
[![Build Status](https://img.shields.io/github/actions/workflow/status/blueprintue/discord-bot/build?label=build&logo=github)](https://github.com/blueprintue/discord-bot/actions?query=workflow%3Abuild)
[![Docker Stars](https://img.shields.io/docker/stars/blueprintue/discord-bot?logo=docker)](https://hub.docker.com/r/blueprintue/discord-bot/)
[![Docker Pulls](https://img.shields.io/docker/pulls/blueprintue/discord-bot?logo=docker)](https://hub.docker.com/r/blueprintue/discord-bot/)
[![Go Report Card](https://goreportcard.com/badge/github.com/blueprintue/discord-bot)](https://goreportcard.com/report/github.com/blueprintue/discord-bot)

## Usage

### From binary

`discord-bot` binaries are available on [releases page](https://github.com/blueprintue/discord-bot/releases/latest).

Choose the archive matching the destination platform:

```shell
wget -qO- https://github.com/blueprintue/discord-bot/releases/download/v0.1.0/discord-bot_0.1.0_linux_arm64.tar.gz | tar -zxvf - discord-bot
```

### From Dockerfile

| Registry                                                                                                  | Image                             |
|-----------------------------------------------------------------------------------------------------------|-----------------------------------|
| [Docker Hub](https://hub.docker.com/r/blueprintue/discord-bot/)                                           | `blueprintue/discord-bot`         |
| [GitHub Container Registry](https://github.com/users/blueprintue/packages/container/package/discord-bot)  | `ghcr.io/blueprintue/discord-bot` |

Following platforms for this image are available:

```
$ docker buildx imagetools inspect blueprintue/discord-bot:latest
Name:      docker.io/blueprintue/discord-bot:edge
MediaType: application/vnd.docker.distribution.manifest.list.v2+json
Digest:    sha256:8d4501ea22b8914315a26acbf3da17c1009d15e480a80a75f1d2166db48ac998

Manifests:
  Name:      docker.io/blueprintue/discord-bot:edge@sha256:7194ed60cdcd51c57389c666d7e325464486e6fb7f593c3db57230ff0f05c40b
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/amd64

  Name:      docker.io/blueprintue/discord-bot:edge@sha256:7c984e4b0a0f9e106d201e15633e17319e4a312d283a16fa7b2782b6ddc9bb57
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/arm/v7

  Name:      docker.io/blueprintue/discord-bot:edge@sha256:329c01304b303d64016ceffdfa1b1c50731212d72db5e070f04923f0375f4df4
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/arm64
```

## Build

```shell
git clone https://github.com/blueprintue/discord-bot.git discord-bot
cd discord-bot

# build docker image and output to docker with discord-bot tag (default)
docker buildx bake

# build multi-platform image
docker buildx bake image-all

# create artifacts in ./dist
docker buildx bake artifact-all
```