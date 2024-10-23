ARG BUILDER=golang:1.18
ARG BASE=alpine:3.10

FROM ${BUILDER} as builder
ARG GOPROXY=direct

WORKDIR /workspace
COPY . .

RUN go mod download
RUN GOPROXY=${GOPROXY} CGO_ENABLED=0 go build -ldflags "-w -s" -o argo-wf-atomic

FROM ${BASE}

COPY --from=builder /workspace/argo-wf-atomic /usr/bin/argo-wf-atomic
RUN apk add ca-certificates
CMD ["argo-wf-atomic"]
