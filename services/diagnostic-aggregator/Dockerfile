# syntax=docker/dockerfile:1

FROM golang:1.26.3-alpine AS build

WORKDIR /src

COPY services/diagnostic-aggregator/go.mod services/diagnostic-aggregator/go.sum ./services/diagnostic-aggregator/
COPY shared/go/observabilityevent/go.mod ./shared/go/observabilityevent/
COPY shared/go/servicelog/go.mod ./shared/go/servicelog/
RUN cd services/diagnostic-aggregator && go mod download

COPY services/diagnostic-aggregator ./services/diagnostic-aggregator
COPY shared/go/observabilityevent ./shared/go/observabilityevent
COPY shared/go/servicelog ./shared/go/servicelog
RUN cd services/diagnostic-aggregator && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/diagnostic-aggregator ./cmd/diagnostic-aggregator
RUN mkdir -p /out/data/reports /out/data/logs && chown -R 65532:65532 /out/data

FROM scratch

COPY --from=build /out/diagnostic-aggregator /diagnostic-aggregator
COPY --from=build --chown=65532:65532 /out/data /data

ENV PORT=8080 \
    ENVIRONMENT=production \
    DIAGNOSTIC_AGGREGATOR_ENABLED=true \
    DIAGNOSTIC_AGGREGATOR_STORAGE_ROOT=/data/reports \
    DIAGNOSTIC_AGGREGATOR_LOG_ROOT=/data/logs

VOLUME ["/data"]
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/diagnostic-aggregator"]
