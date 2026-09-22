FROM golang:1.25 AS builder

WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend .

RUN apt-get update && apt-get install -y --no-install-recommends curl jq && rm -rf /var/lib/apt/lists/*

ARG FRANKFURTER_OPENAPI_URL=https://api.frankfurter.dev/v2/openapi.json
ENV FRANKFURTER_OPENAPI_URL=$FRANKFURTER_OPENAPI_URL

RUN go generate ./...
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/currency-watcher ./cmd/server

FROM gcr.io/distroless/base-debian12:nonroot
COPY --from=builder /out/currency-watcher /currency-watcher
EXPOSE 8080
ENTRYPOINT ["/currency-watcher"]
