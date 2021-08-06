FROM golang:1.16-buster as builder

WORKDIR /tmp/go

COPY go.* .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./social_auth .

FROM alpine:3.14.0
RUN apk add ca-certificates

COPY --from=builder /tmp/go/social_auth /social_auth
COPY resources resources

EXPOSE 443

CMD ["/social_auth"]
