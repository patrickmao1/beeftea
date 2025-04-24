FROM golang:1.24.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ./node ./cmd/beeftea

FROM alpine:latest

WORKDIR /app

RUN mkdir runtime

COPY --from=builder /app/node .

CMD ./node 2>&1 | tee ./runtime/app.log