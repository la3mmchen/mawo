# build
FROM golang:alpine as builder
ARG app_version
RUN mkdir /build
ADD . /build/
WORKDIR /build

RUN go build -o main -ldflags "-X main.AppVersion=$app_version" .

# run
FROM alpine:3
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/main /app/
WORKDIR /app
ENTRYPOINT [ "./main" ]
CMD [""]