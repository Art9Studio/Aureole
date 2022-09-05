FROM golang:1.18.3-buster as builder

WORKDIR /tmp/go

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./aureole .

FROM alpine:3.16.0
RUN apk add ca-certificates

COPY --from=builder /tmp/go/aureole /aureole

EXPOSE 3000

CMD ["/aureole"]
