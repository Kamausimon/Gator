# Gator

Gator is a command-line RSS feed aggregator. It lets multiple users register, follow RSS feeds, continuously scrape new posts in the background, and browse the posts from feeds they follow — all from the terminal.

## Requirements

Before installing Gator, make sure you have the following installed:

- [Go](https://go.dev/doc/install) (1.25 or later)
- [PostgreSQL](https://www.postgresql.org/download/)

## Installing

Install the `gator` CLI using `go install`:

```bash
go install github.com/Kamausimon/gator@latest
```

This builds the binary and places it in your `$GOPATH/bin` (or `$HOME/go/bin`), so make sure that directory is on your `PATH`.

## Configuration

Gator reads its configuration from a JSON file at `~/.gatorconfig.json`. Create it with the following structure:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

- `db_url` — connection string for your Postgres database. Make sure the `gator` database exists and the schema migrations have been applied.
- `current_user_name` — leave this blank initially; it's managed automatically by the `login` and `register` commands.

## Running Gator

Once installed and configured, run commands using:

```bash
gator <command> [arguments]
```

### Available commands

- `register <name>` — create a new user and log in as them.
- `login <name>` — switch the current user.
- `users` — list all registered users, marking the currently logged-in one.
- `reset` — delete all users (and their data) from the database.
- `addfeed <name> <url>` — add a new RSS feed and automatically follow it.
- `feeds` — list all feeds registered by any user.
- `follow <url>` — follow an existing feed as the current user.
- `following` — list the feeds the current user is following.
- `unfollow <url>` — stop following a feed.
- `agg <time_between_reqs>` — start a long-running aggregator that periodically fetches new posts from the least-recently-fetched feed (e.g. `gator agg 1m`). Stop it with `Ctrl+C`.
- `browse [limit]` — show the most recent posts from feeds the current user follows (defaults to 2 posts).

### Example workflow

```bash
gator register alice
gator addfeed "Hacker News" "https://hnrss.org/newest"

# In a separate terminal, leave this running to collect posts
gator agg 1m

# Back in your first terminal
gator browse 5
```
