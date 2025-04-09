# stserver


## How to run migrations

```
goose -dir ./db/migrations postgres "postgres://dev:devpass@localhost:5432/sovereign?sslmode=disable" up
```

## How to generate new typesafe code from SQL

In the db dir
```
sqlc generate
```