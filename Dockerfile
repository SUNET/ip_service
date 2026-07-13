FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /go/src/app

# Install swagger tool (cached unless Go version changes)
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy dependency files first for layer caching
COPY go.mod go.sum ./
COPY vendor/ vendor/

# Copy source code and build files
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
COPY docs/ docs/
COPY Makefile .

RUN make swagger

ARG GIT_COMMIT=unknown
ARG GIT_BRANCH=unknown

RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/ip_service -ldflags \
    "-X ip_service/pkg/model.BuildVariableGitCommit=${GIT_COMMIT} \
    -X ip_service/pkg/model.BuildVariableGitBranch=${GIT_BRANCH} \
    -X ip_service/pkg/model.BuildVariableTimestamp=$(date +'%F:T%TZ') \
    -X ip_service/pkg/model.BuildVariableGoVersion=$(go version|awk '{print $3}') \
    -X ip_service/pkg/model.BuildVariableGoArch=$(go version|awk '{print $4}') \
    -w -s" ./cmd/ip_service/main.go

## Deploy
FROM alpine:3.21

RUN apk add --no-cache curl procps iputils less

WORKDIR /

COPY --from=builder /go/src/app/bin/ip_service /ip_service

EXPOSE 8080

HEALTHCHECK --interval=20s --timeout=10s CMD curl --connect-timeout 5 http://localhost:8080/health | grep -q STATUS_OK

CMD [ "./ip_service" ]