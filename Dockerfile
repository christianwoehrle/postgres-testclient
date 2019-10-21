FROM alpine:latest
RUN apk --update add postgresql-client && rm -rf /var/cache/apk/*
COPY testmaster.sh /testmaster.sh
ENTRYPOINT [ "psql" ]
