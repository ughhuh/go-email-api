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
        "compress": false
    }
}
```

- `mode` is the Gin mode. Available values: `debug`, `test`, `release`.
- `port` is the port to listen on for API requests.
- `ssl_mode` is the database SSL mode. Available values: `disable`, `allow`, `prefer`, `require`, `verify-ca`, `verify-full`.
- `allowed_domains` is the list of domains that the API manages emails for.
- `trusted_proxies` is the list of IP addresses of proxy servers that API can trust to accurately forward information about the original client's request.
- `logrotate` is the list of configurations for the log rolling
  - `log_file` is the file to write logs to. Backup logs will be retained in the same directory.
  - `max_size` is the maximum file size in Mb before it gets rotated
  - `max_backups` is the maximum number of backup files to retain
  - `max_age` is the maximum number of days to retain the backup files for
  - `compress` is the flag indicating whether the rotated log files should be compressed. Compression is done with `gzip`.

Database connection parameters are defined as environmental variables.

`.env`:

```sh
DB_HOST=string
DB_NAME=string
DB_USER=string
DB_SECRET=string
```

- `DB_HOST` is the database hostname. Can be set to `localhost`, `host.docker.internal`, or the name of the service when used with Docker compose.
- `DB_NAME` is the name of the database.
- `DB_USER` is the database user to use to execute queries.
- `DB_SECRET` is the password of the database user.

## Endpoints

### `GET /inbox/:address`

Fetches the contents of the existing inbox. Returns a list of emails.

Request parameters:

- `address` is the email address to fetch emails for. Format: `username@domain.tld`.

Response body:

```json
{
    "success": true,
    "emails": [
        {
            "message_id": "string",
            "from": "string",
            "date": "string"
        }
    ]
}
```

- `success` indicates the operation status
- `emails` is the list of emails fetched for the user
  - `message_id` is the value of the `Message-Id` header of the email message.
  - `from` is the value of the `From` header of the email message.
  - `date` is the datetime timestamp of when the email message was received. Format: `YYYY-MM-DDThh:mm:ss.SSSZ` (ISO 8601 datetime timestamp)

The data is fetched using `getSimpleEmailByMsgId` query that is parsed to the `SimpleEmail` struct.

Returns status code 200 on success. Returns an empty `emails` list if the user doesn't have any emails addressed to them.

Errors:

- `500: Failed to retrieve emails` when the API fails to fetch emails
- `500: An error occured while parsing email records` when the API fails to parse the data retrieved from the emails table.

### `GET /email/:message_id`

Fetches the email by its `message_id`. Returns the headers and body of the email message.

Response body:

```json
{
    "message_id": "string",
    "body": "string",
    "from": ["string"],
    "to": ["string"]
}
```

- `message_id` is the value of the `Message-Id` header of the email message.
- `body` is the body part of the email message.
- `from` is the value of the `From` header of the email message.
- `to` is the value of the `To` header of the email message.

The data is fetched using `getEmailByMsgId` query that is parsed to the `Email` struct.

Returns status code 200 on success.

Errors:

- `500: Failed to retrieve emails` when the API fails to fetch emails
- `500: An error occured while parsing email records` when the API fails to parse the data retrieved from the emails table.
- `404: Email not found` when the email with `message_id` doesn't exist in the database.

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
- `domain` is the desired domain. Domain must be present in the `allowed_domains` configuration setting list. Mandatory, set to "" to get a random domain.

The payload format is controlled by the `CreateEmailRequest` struct.

Response body:

```json
{
    "success": true,
    "email_address": "string"
}
```

- `success` indicates the operation status
- `email_address` is the created email address. Format: `username@domain.tld`

The data is inserted using `createNewUser` query.

Returns status code 201 on success.

Errors:

- `400: Invalid request format` when the request payload doesn't match the expected one.
- `400: Invalid domain name` when the `domain` is not present in the `allowed_domains` list set in the configuration file.
- `500: Failed to create new email address` when the API fails to create a new user.

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

- `success` indicates the operation status
- `error` is the field that contains an error message if the operation failed

The endpoint fetched the list of the messages per user with `getMsgIdsByUserId` query, deletes the user with `deleteUserByUserId`, and deletes the messages linked to the user with `deleteMsgByMsgId`.
`getMsgIdsByUserId` query is designed to ensure that the message ids returned belong only to the target user to avoid accidentally removing emails that are addressed to other users too.

Returns status code 200 on success.

Errors:

- `400: Invalid request format` when the request payload doesn't match the expected one.
- `400: Invalid email address` when the `email_address` field is left empty.
- `500: An error occured while parsing email records` when the API fails to parse the inbox entries for the user. API doesn't fail if there aren't any emails associated with the email address.
- `500: Failed to delete email user` when the `deleteUserByUserId` query fails.
- `500: Failed to delete user's emails` when the `deleteMsgByMsgId` query fails.

## Improvement ideas

- add a check to the post new email to ensure the email is compliant with RFC standards for email usernames
- move db ping time to configuration, make db ping optional
- not sure if hardcoded tables and columns in the queries is a good idea

todo add doc of env vars
