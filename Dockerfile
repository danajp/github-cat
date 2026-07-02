FROM golang:1.24 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd

RUN CGO_ENABLED=0 go build -o /github-cat ./cmd/github-cat

FROM gcr.io/distroless/static-debian12

COPY --from=build /github-cat /github-cat

ENTRYPOINT ["/github-cat"]
