FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY kubeav /usr/local/bin/kubeav
USER 65532:65532

ENTRYPOINT ["/usr/local/bin/kubeav"]
