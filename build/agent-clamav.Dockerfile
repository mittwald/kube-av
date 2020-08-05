FROM golang:1.14 AS builder

WORKDIR /work
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o kubeav-agent -a cmd/agent/main.go

FROM alpine

RUN apk add -U clamav clamav-libunrar

COPY --from=builder /work/kubeav-agent /usr/bin/kubeav-agent

ENTRYPOINT ["/usr/bin/kubeav-agent"]