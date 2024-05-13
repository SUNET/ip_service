FROM golang:latest AS builder

WORKDIR /go/src/app

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build GOOS=linux GOARCH=amd64 go build -v -o bin/ip_service -ldflags \
    "-X ip_service/pkg/model.BuildVariableGitCommit=$(git rev-list -1 HEAD) \
    -X ip_service/pkg/model.BuildVariableGitBranch=$(git rev-parse --abbrev-ref HEAD) \
    -X ip_service/pkg/model.BuildVariableTimestamp=$(date +'%F:T%TZ') \
    -X ip_service/pkg/model.BuildVariableGoVersion=$(go version|awk '{print $3}') \
    -X ip_service/pkg/model.BuildVariableGoArch=$(go version|awk '{print $4}') \
    -w -s --extldflags '-static'" ./cmd/main.go

## Deploy
FROM debian:bookworm-slim

WORKDIR /

RUN apt-get update && apt-get install -y curl procps iputils-ping less
RUN rm -rf /var/lib/apt/lists/*

COPY --from=builder /go/src/app/bin/ip_service /ip_service

EXPOSE 8080

HEALTHCHECK --interval=20s --timeout=10s CMD curl --connect-timeout 5 http://localhost:8080/health | grep -q STATUS_OK

CMD [ "./ip_service" ]