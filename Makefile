.PHONY: update clean build build-all run package deploy test authors dist

NAME 					:= ip_service
LDFLAGS                 := -ldflags "-w -s --extldflags '-static'"

gosec:
	$(info Run gosec)
	gosec -color -nosec -tests ./...

staticcheck:
	$(info Run staticcheck)
	staticcheck ./...

start:
	$(info Run!)
	docker-compose -f docker-compose.yaml up -d --remove-orphans

stop:
	$(info stopping VC)
	docker-compose -f docker-compose.yaml rm -s -f

restart: stop start

ifndef VERSION
VERSION := latest
endif

DOCKER_TAG_IP_SERVICE 		:= docker.sunet.se/ip_service:$(VERSION)

docker-build-ip_service:
	$(info Docker Building ip_service with tag: $(VERSION))
	docker build --tag $(DOCKER_TAG_IP_SERVICE) .

vscode:
	$(info Install APT packages)
	sudo apt-get update && sudo apt-get install -y \
		protobuf-compiler \
		netcat-openbsd
	$(info Install go packages)
	go install golang.org/x/tools/cmd/deadcode@latest && \
	go install github.com/securego/gosec/v2/cmd/gosec@latest && \
	go install honnef.co/go/tools/cmd/staticcheck@latest