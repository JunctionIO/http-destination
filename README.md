# Junction HTTP Destination Worker

Delivers webhook payloads via HTTP POST and publishes delivery status to `junction.status` for the system worker to report back to the API.

## How it fits

On startup the worker registers itself with the Junction API, which upserts the `destination_types` record and declares the `junction.destinations.http` queue. The worker then consumes from that queue — one message per delivery task. For each message it POSTs the payload to the configured destination URL, then publishes a status message (`dispatched` or `errored`) to `junction.status`. The system worker picks that up and calls the API to update the `event_log_destinations` record.

## Environment variables

| Variable | Description |
|---|---|
| `RABBITMQ_URL` | RabbitMQ connection string (`amqp://user:pass@host:5672/`) |
| `JUNCTION_API_URL` | Base URL of the Junction API (`http://api:8080`) |
| `JUNCTION_WORKER_TOKEN` | Worker JWT — generate with the Junction CLI |

Copy `.env.example` to `.env` and fill in the values before running locally.

## Local development

Requires the Junction API devenv to be running (provides RabbitMQ and the registration endpoint). Then:

```
cp .env.example .env
# fill in JUNCTION_WORKER_TOKEN
nix develop
go run .
```

Or start with devenv:

```
devenv up
```

## Testing

```
go test ./...
```

## Destination config

HTTP destinations are configured with the following fields:

| Field | Required | Description |
|---|---|---|
| `url` | Yes | Target endpoint |
| `auth_header` | No | Header name for authentication (e.g. `Authorization`) |
| `auth_token` | No | Header value (e.g. `Bearer abc123`) |

Both `auth_header` and `auth_token` must be set for the auth header to be sent.

## Delivery behaviour

Each message is attempted once with a 30-second timeout. On success (2xx response) a `dispatched` status is published. On failure (non-2xx or network error) an `errored` status is published with the HTTP status code and response body, or the network error message. The delivery task is always acknowledged — the stuck `event_log_destinations` record (status remains `pending`) is the recovery artifact and can be replayed manually.

Malformed envelopes and messages with an unexpected destination type are acknowledged and discarded without publishing a status message.
