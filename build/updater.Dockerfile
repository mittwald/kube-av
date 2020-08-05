FROM alpine

RUN apk add -U clamav
ENTRYPOINT ["freshclam", "-F", "-c", "24", "-d"]