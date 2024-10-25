# Go Email API
Email Management API

available gin modes: debug release test

## Commands

### Docker

Build Docker image:

```bash
docker build --tag docker-email-api .
```

Run Docker container:

```bash
docker run docker-email-api --config=config.json --logdir=./logs
```

API listens on port 80 by default. Connect using `--network=host.docker.internal` for host machine localhost and `--network=network-name` for Docker bridge network.

Map ports with `--publish 8080:80`.
