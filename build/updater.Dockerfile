FROM alpine

RUN apk add -U clamav clamav-libunrar
ENTRYPOINT ["freshclam", "-F", "-c", "24", "-d"]