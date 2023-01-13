build: fmt
	CGO_ENABLE=0 go build -ldflags "-w -s" -o bin/argo-wf-atomic

run:
	go run main.go

fmt:
	go fmt ./...

copy: build
	sudo cp bin/argo-wf-atomic /usr/local/bin/argo-wf-atomic

image:
	docker build . -t ghcr.io/linuxsuren/argo-workflow-atomic-plugin:master
