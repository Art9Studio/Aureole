FROM golang:1.18.1-buster as builder

WORKDIR /tmp/go/

COPY src/go.mod .
COPY src/go.sum .

RUN go mod download

COPY src .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./aureole .

FROM alpine:3.15.4
RUN apk add ca-certificates

COPY --from=builder /tmp/go/aureole /aureole

EXPOSE 3000

CMD ["/aureole"]
