# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26.3

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

COPY services/game-server/go.mod services/game-server/go.sum services/game-server/
COPY services/player-data/go.mod services/player-data/go.sum services/player-data/
COPY services/diagnostic-aggregator/go.mod services/diagnostic-aggregator/go.sum services/diagnostic-aggregator/
COPY shared/go/servicelog/go.mod shared/go/servicelog/
COPY shared/go/observabilityevent/go.mod shared/go/observabilityevent/

RUN cd services/game-server && go mod download

COPY services/game-server services/game-server
COPY services/player-data services/player-data
COPY services/diagnostic-aggregator services/diagnostic-aggregator
COPY shared/go shared/go

RUN mkdir -p \
        /out/app/data/diagnostic-reports \
        /out/app/logs/diagnostic-aggregator \
        /out/app/logs/game-server \
        /out/app/logs/player-data \
        /out/app/measurement-results/game-server \
    && cd services/game-server \
    && CGO_ENABLED=0 GOOS=linux go build \
        -tags=nodevtools,noembeddedsqlite \
        -trimpath \
        -buildvcs=false \
        -ldflags='-s -w' \
        -o /out/app/game-server \
        ./cmd/game-server

COPY shared/collisions/collision_shapes.json /out/app/shared/collisions/collision_shapes.json

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build --chown=nonroot:nonroot /out/app /app

WORKDIR /app
USER nonroot:nonroot

EXPOSE 8080/tcp
STOPSIGNAL SIGTERM

ENTRYPOINT ["/app/game-server"]
