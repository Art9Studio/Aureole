FROM golang:1.17.2-buster as builder

WORKDIR /tmp/go

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./aureole .

FROM alpine:3.14.3
RUN apk add ca-certificates

COPY --from=builder /tmp/go/aureole /aureole

EXPOSE 3000

CMD ["/aureole"]
