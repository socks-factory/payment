FROM registry.opensuse.org/opensuse/bci/golang:1.23 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . ./

RUN make

FROM scratch

COPY --from=build /app/payment /payment

LABEL org.label-schema.vendor="SUSE" \
  org.label-schema.build-date="${BUILD_DATE}" \
  org.label-schema.version="${BUILD_VERSION}" \
  org.label-schema.name="Socks Shop: Payment" \
  org.label-schema.description="REST API for Payment service" \
  org.label-schema.url="https://github.com/socks-factory/payment" \
  org.label-schema.vcs-url="github.com:socks-factory/payment.git" \
  org.label-schema.vcs-ref="${COMMIT}" \
  org.label-schema.schema-version="1.0"

EXPOSE 8080
CMD ["/payment", "-port=8080"]

