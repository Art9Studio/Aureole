FROM golang:1.19.3-buster as builder

WORKDIR /tmp/go

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY src .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./aureole .

FROM postgres:alpine3.14

COPY --from=builder /tmp/go/aureole /app/aureole
COPY deployments/render /app/render

RUN mv /app/render/templates /app/render/res
RUN mv /app/render/entrypoint.sh /app
RUN mv /app/render/config.yaml /app

CMD ["/app/entrypoint.sh"]