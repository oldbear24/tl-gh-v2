# TL-GH v2

TL-GH v2 is a Discord bot written in Go. It connects to a PostgreSQL
database, registers a set of application commands and listens for events
from the Discord gateway. The bot includes helper utilities such as dice
rolling and gear lookups for guild members.

## Configuration

The application reads configuration from environment variables prefixed
with `TLGH_` as well as from command‑line flags. Key variables include:

- `TLGH_BOT_TOKEN` – Discord bot token.
- `TLGH_POSTGRE_CONN_STRING` – PostgreSQL connection string.
- `TLGH_EXTLOG_ENABLE` – enables external logging when set to `true`.
- `TLGH_LOG_URL` / `TLGH_LOG_TOKEN` / `TLGH_LOG_ORG` / `TLGH_LOG_STREAM` –
  parameters for the external log sink.
- `TLGH_LOG_LEVEL` – logging level (`-4` debug, `0` info, `4` warn,
  `8` error).

Flags with matching names can be used as an alternative way to specify the
same settings.

## Development

Ensure you have Go installed, then fetch dependencies and run tests:

```bash
go mod tidy
go test ./...
```

## Running

Set the required configuration values and start the bot:

```bash
export TLGH_BOT_TOKEN=your_token_here
export TLGH_POSTGRE_CONN_STRING="postgres://user:pass@localhost/dbname"
go run .
```

The bot will connect to Discord, register its commands and begin processing
events.

