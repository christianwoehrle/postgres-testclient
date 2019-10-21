FROM alpine:latest
RUN apk --update add postgresql-client && rm -rf /var/cache/apk/*
COPY testmaster.sh /test.sh
ENTRYPOINT [ "psql" ]
