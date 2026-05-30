FROM golang:1.25.3-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o app ./cmd


FROM alpine:3.21
WORKDIR /app
COPY --from=builder /build/app .
EXPOSE 8080
CMD [ "./app" ]
