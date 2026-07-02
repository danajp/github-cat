FROM golang:1.24 AS build

WORKDIR /app

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY cmd ./cmd

RUN CGO_ENABLED=0 go build -mod=vendor -o /github-cat ./cmd/github-cat

FROM gcr.io/distroless/static-debian12

COPY --from=build /github-cat /github-cat

ENTRYPOINT ["/github-cat"]
