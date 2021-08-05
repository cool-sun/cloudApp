FROM  golang:1.17rc1-alpine AS builder
WORKDIR /cloud-app
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -mod=vendor

FROM alpine:3.14.0
WORKDIR /root/
COPY --from=builder /cloud-app/cloud-app .
COPY --from=builder /cloud-app/static .
COPY --from=builder /cloud-app/view .
COPY --from=builder /cloud-app/template .
CMD ["./cloud-app"]