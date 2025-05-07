FROM registry.opensuse.org/opensuse/bci/golang:1.23 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . ./

RUN make

FROM alpine:3

COPY --from=build /app/payment /payment

ENV	SERVICE_USER=myuser \
	SERVICE_UID=10001 \
	SERVICE_GROUP=mygroup \
	SERVICE_GID=10001

RUN	addgroup -g ${SERVICE_GID} ${SERVICE_GROUP} && \
	adduser -g "${SERVICE_NAME} user" -D -H -G ${SERVICE_GROUP} -s /sbin/nologin -u ${SERVICE_UID} ${SERVICE_USER} && \
	chmod +x /payment && \
    chown -R ${SERVICE_USER}:${SERVICE_GROUP} /payment

LABEL org.label-schema.vendor="SUSE" \
  org.label-schema.build-date="${BUILD_DATE}" \
  org.label-schema.version="${BUILD_VERSION}" \
  org.label-schema.name="Socks Shop: Payment" \
  org.label-schema.description="REST API for Payment service" \
  org.label-schema.url="https://github.com/socks-factory/payment" \
  org.label-schema.vcs-url="github.com:socks-factory/payment.git" \
  org.label-schema.vcs-ref="${COMMIT}" \
  org.label-schema.schema-version="1.0"

USER ${SERVICE_USER}

EXPOSE 80
CMD ["/payment", "-port=80"]

