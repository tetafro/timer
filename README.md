# timer

[![Go Report](https://goreportcard.com/badge/github.com/tetafro/timer)](https://goreportcard.com/report/github.com/tetafro/timer)
[![CI](https://github.com/tetafro/timer/actions/workflows/push.yml/badge.svg)](https://github.com/tetafro/timer/actions)

Simple web form for creating timers.

[Live version](https://timer.dkrv.me).

## Run

Compile and run binary
```sh
make build run
```

Run as a docker container
```sh
docker run -d ghcr.io/tetafro/timer
```

## Deploy

```sh
SSH_SERVER=164.90.195.10:15222 \
SSH_USER=tetafro \
make deploy
```
