# Go-ShortUrl

> An url shortener written in Go and uses Redis as database.

## Build

```bash
docker build . go-shorturl
```

## Run

Copy `.env.example` to `.env` and change the values to your own.

To enable basic auth on APIs, set `ENABLE_BASIC_AUTH=true` in `.env`.
Do not forget to set a username and password.

Finally.

```bash
docker-compose up -d
```
