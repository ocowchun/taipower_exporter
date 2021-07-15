FROM golang:1.16 as builder
WORKDIR /go/src/github.com/ocowchun/taipower_exporter
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY collector ./collector
RUN CGO_ENABLED=0 GOOS=linux go build -a -o taipower_exporter .


FROM alpine:latest
RUN apk --no-cache add ca-certificates
USER 1001:1001
COPY --from=builder /go/src/github.com/ocowchun/taipower_exporter/taipower_exporter /usr/bin/
ENTRYPOINT [ "/usr/bin/taipower_exporter" ]
