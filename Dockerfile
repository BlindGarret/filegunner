FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY . .
RUN go mod download
RUN go mod verify
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /main ./cmd

FROM alpine 
COPY --from=builder main .
CMD ["./main"]