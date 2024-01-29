## Discord-bot

[![GitHub release](https://img.shields.io/github/release/blueprintue/discord-bot.svg?logo=github)](https://github.com/blueprintue/discord-bot/releases/latest)
[![Total downloads](https://img.shields.io/github/downloads/blueprintue/discord-bot/total.svg?logo=github)](https://github.com/blueprintue/discord-bot/releases/latest)
[![Build Status](https://img.shields.io/github/actions/workflow/status/blueprintue/discord-bot/build.yml?label=build&logo=github)](https://github.com/blueprintue/discord-bot/actions/workflows/build.yml)
[![Docker Stars](https://img.shields.io/docker/stars/blueprintue/discord-bot?logo=docker)](https://hub.docker.com/r/blueprintue/discord-bot/)
[![Docker Pulls](https://img.shields.io/docker/pulls/blueprintue/discord-bot?logo=docker)](https://hub.docker.com/r/blueprintue/discord-bot/)
[![Go Report Card](https://goreportcard.com/badge/github.com/blueprintue/discord-bot)](https://goreportcard.com/report/github.com/blueprintue/discord-bot)

## Usage

### From binary

`discord-bot` binaries are available on [releases page](https://github.com/blueprintue/discord-bot/releases/latest).

### From Dockerfile

| Registry                                                                                                  | Image                             |
|-----------------------------------------------------------------------------------------------------------|-----------------------------------|
| [Docker Hub](https://hub.docker.com/r/blueprintue/discord-bot/)                                           | `blueprintue/discord-bot`         |
| [GitHub Container Registry](https://github.com/users/blueprintue/packages/container/package/discord-bot)  | `ghcr.io/blueprintue/discord-bot` |

## Build from source

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

## Run docker image inside project
If you git clone you will see a `docker-composer.yml` already setup for local image.  
You need to create `config.json` in project folder.  
By default log folder is `/var/log/discord-bot` so you need to set:  
```json
    "log": {
      "filename": "/var/log/discord-bot/discord-bot.log",
      "level": "info"
    },
```
Then you can run image.
```shell
docker-compose up --build -d
```
If you want to use image from repository replace `image` in `docker-composer.yml`.

## Configuration explanations
### General
Mandatory parameters to run discord-bot without modules.

#### Discord
| JSON Parameter | ENV Parameter      | Mandatory | Type   | Description                                                |
| -------------- | ------------------ | --------- | ------ | ---------------------------------------------------------- |
| name           | DBOT_DISCORD_NAME  | YES       | string | discord server name (also called guild name)               |
| token          | DBOT_DISCORD_TOKEN | YES       | string | token for a bot                                            |

##### How to get discord name?
When you are on a discord server, you will see a list of channels on the left, at the top you will see the discord server name.

##### How to generate token?
You need to create a bot, you can start by looking at tutorial from Discord: [https://discord.com/developers/docs/getting-started](https://discord.com/developers/docs/getting-started).

#### Log
| JSON Parameter | ENV Parameter      | Mandatory | Type   | Specific values                                           | Description                                                                     |
| -------------- | ------------------ | --------- | ------ | --------------------------------------------------------- | ------------------------------------------------------------------------------- |
| filename       | DBOT_LOG_FILENAME  | YES       | string |                                                           | relative or absolute path to log file (it will create directories if not exist) |
| level          | DBOT_LOG_LEVEL     | NO        | string | trace \| debug \| info \| warn \| error \| fatal \| panic | level of log (if empty then no log)                                             |

##### What is level?
It uses zerolog levels (from highest to lowest):
* panic (zerolog.PanicLevel, 5)
* fatal (zerolog.FatalLevel, 4)
* error (zerolog.ErrorLevel, 3)
* warn (zerolog.WarnLevel, 2)
* info (zerolog.InfoLevel, 1)
* debug (zerolog.DebugLevel, 0)
* trace (zerolog.TraceLevel, -1)

### Modules
#### Welcome
Define the user's role when using an emoji.  
You can define one or more messages in only one channel.  
In each message you can define title, description, color, role and emoji to use.  

You can see an example below:  
![Welcome module example](doc/welcome-example.png)

JSON configuration used:  
```json
"welcome": {
  "channel": "😃test",
  "messages": [
    {
      "title": "Are you interested by blueprintUE?",
      "description": "Click the :blueprintUE: below to join the blueprintUE part",
      "color": 5301186,
      "role": "is member",
      "emoji": "blueprintUE"
    }
  ]
}
```

##### Channel
You can define only one channel.  

| JSON Parameter | Mandatory | Type   | Description  |
| -------------- | --------- | ------ | ------------ |
| channel        | YES       | string | channel name |

##### Message
You can defines multiple messages.  

| JSON Parameter | Mandatory | Type   | Default value | Description                                                                                   |
| -------------- | --------- | ------ | ------------- | --------------------------------------------------------------------------------------------- |
| title          | YES       | string |               | title's message                                                                               |
| description    | YES       | string |               | description's message                                                                         |
| color          | NO        | int    | 0             | color on the left of the message (format is integer representation of hexadecimal color code) |
| role           | YES       | string |               | role's name to assign when user use correct emoji                                             |
| emoji          | YES       | string |               | emoji to use (format is my_emoji without `:`)                                                 |

##### How it works?
Each time you start `discord-bot`, welcome module will check the configuration in the `config.json`.  
If there is nothing missing, it will fetch channels, roles and emoji.  
Then it will do another check to see if channel, role and emoji exists.  

Secondly it will listen two events on `onMessageReactionAdd` and `onMessageReactionRemove`.  

After it will search the message in the channel.  
If the message is not found then it will publish it and add a reaction to show user which emoji to use.  
If the message is found then it will fetch all reactions by the users and apply role.