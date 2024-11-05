# Go Email API
Email Management API

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

API listens on port 8080 by default. Use `host.docker.internal` as `DB_HOST` when connecting to DB container on Windows without Docker compose. Use the name of the service as `DB_HOST` when connecting using Docker compose.

Map ports with `--publish 8080:80`.

## Configuration

```json
{
    "mode": "debug",
    "port": 8080,
    "ssl_mode": "disable",
    "allowed_domains": [
        "vm4408.kaj.pouta.csc.fi"
    ],
    "trusted_proxies": [
        "127.0.0.1"
    ],
    "logrotate": {
        "log_file": "api.log",
        "max_size": 10,
        "max_backups": 3,
        "max_age": 28,
        "compress": true
    }
}
```

- `mode` is the Gin mode. Available values: `debug`, `test`, `release`.
- `port` is the port to listen on for API requests.
- `ssl_mode` is the database SSL mode. Available values: `disable`, `allow`, `prefer`, `require`, `verify-ca`, `verify-full`.
- ``

## Endpoints

### `GET /inbox/:email_id`

Fetches the contents of the existing inbox. Returns a list of emails.

Example response:

```json
{
    "success": true,
    "emails": [
        {
            "message_id": "string",
            "from": "string",
            "date": "datetime"
        }
    ]
}
```

### `GET /email/:email_id`

Fetches the email by its `message_id`. Returns the headers and body of the email message.

Response:

```json
{
    "message_id": "string",
    "body": "string",
    "from": ["string"],
    "to": ["string"]
}

```

### `POST /email`

Creates a new inbox.

Request payload:

```json
{
    "username": "string",
    "domain": "string"
}
```

- `username` is the desired username. Mandatory, set to "" to get a random username.
- `domain` is the desired domain. Mandatory, set to "" to get a random domain.

Response:

```json
{
    "success": true,
    "email_address": "string"
}
```

- `email_address` is the created email address.

### `DELETE /email`

Deletes the inbox.

Request payload:

```json
{
    "email_address": "string"
}
```

- `email_address` is the email address to delete.

Response:

```json
{
    "success": true,
    "error": "string"
}
```
