ARG BUILDER=golang:1.18
ARG BASE=ubuntu:kinetic

FROM ${BUILDER} as builder
ARG GOPROXY=direct

WORKDIR /workspace
COPY . .

RUN go mod download
RUN GOPROXY=${GOPROXY} CGO_ENABLE=0 go build -ldflags "-w -s" -o argo-wf-atomic

FROM ${BASE}

COPY --from=builder /workspace/argo-wf-atomic /usr/bin/argo-wf-atomic
RUN apt update -y && apt install ca-certificates -y
CMD ["argo-wf-atomic"]
