FROM golang:1.18 AS build

RUN useradd -u 10001 benthos

WORKDIR /build/
COPY . /build/

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor

FROM busybox AS package

LABEL maintainer="zgldh <zgldh@hotmail.com>"

WORKDIR /

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /build/benthos_modbus_processor .
COPY ./config/example.yaml /benthos.yaml

RUN mkdir /logs
# RUN chown benthos:10001 benthos_modbus_processor
RUN chown -R benthos:10001 /logs

USER benthos

EXPOSE 4195

ENTRYPOINT ["/benthos_modbus_processor"]

CMD ["-c", "/benthos.yaml"]
