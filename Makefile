all: templates build

templates:
	qtc tpl

build:
	CGO_ENABLED=0 \
	GOOS=linux \
	go build \
		-a \
		-o app.cmd \
		-ldflags "-s -w -extldflags -static" \
		cmd/server/main.go

.PHONY: templates build