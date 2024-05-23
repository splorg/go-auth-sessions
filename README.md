# go-auth-sessions

Example/template repository setting up <ins>session</ins> authentication in Go.

[Check the branch "net/http" for a version of this template made using the standard library instead of Fiber](https://github.com/splorg/go-auth-sessions/tree/net/http)

This project uses:

- [Fiber](https://gofiber.io/) for routing/middleware
- [Goose](https://pressly.github.io/goose/) for database migrations
- [SQLC](https://sqlc.dev/) for generating type-safe code for SQL queries
- [Redis](https://redis.io/) for storing user sessions
- [Postgres](https://www.postgresql.org/) as database