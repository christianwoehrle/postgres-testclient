FROM alpine:latest
USER 1000
RUN apk --update add postgresql-client && rm -rf /var/cache/apk/*
COPY test.sh /test.sh
ENTRYPOINT [ "psql" ]
