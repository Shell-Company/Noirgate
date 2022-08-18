FROM golang:alpine as builder
RUN apk add --no-cache ca-certificates
WORKDIR /build
COPY ./cmd .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -extldflags "-static"' -o noirgate-mgr .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/noirgate-mgr ./noirgate-mgr
ENTRYPOINT ["./noirgate-mgr"]
CMD ["-w","-txt","-v"]