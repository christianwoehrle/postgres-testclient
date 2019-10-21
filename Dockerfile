FROM alpine:latest
RUN apk --update add postgresql-client && rm -rf /var/cache/apk/*
USER 1000
COPY test.sh /test.sh
ENTRYPOINT [ "psql" ]
