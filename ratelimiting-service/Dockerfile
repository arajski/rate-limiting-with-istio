FROM golang:1.24.2-bullseye AS build
WORKDIR /rate-limiting-service

COPY . .

RUN go build -o rate-limiting-service -v main.go

FROM debian:bookworm-slim AS final
RUN apt-get update && apt-get install -y --no-install-recommends sqlite3 && rm -rf /var/lib/apt/lists/*

COPY --from=build /rate-limiting-service/rate-limiting-service /bin/rate-limiting-service
COPY templates/ templates/
COPY static/ static/

ENTRYPOINT [ "/bin/rate-limiting-service" ]
