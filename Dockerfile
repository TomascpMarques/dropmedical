FROM golang:1.22.3-bookworm as builder

WORKDIR /app

COPY . .
RUN go mod download && go mod verify
RUN go build -v -race -o dropmedical

# -----------------------------------

FROM debian:bookworm-slim AS runtime

WORKDIR /app
COPY --from=builder /app/dropmedical /app/dropmedical
COPY .env/prod.env /app/.env/prod.env

ENV ENVIRONMENT="production"

EXPOSE 8081
EXPOSE 1883

CMD [ "./dropmedical" ]