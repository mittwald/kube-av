FROM alpine

RUN apk add -U clamav clamav-libunrar
COPY kubeav-agent /usr/bin/kubeav-agent

ENTRYPOINT ["/usr/bin/kubeav-agent"]