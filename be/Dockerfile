FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY ./ ./

RUN apk add --no-cache gcc musl-dev sqlite-dev
RUN CGO_ENABLED=1 go build -o /go/bin/vego cmd/main.go

FROM alpine:3.21

RUN apk add --no-cache sqlite-libs

WORKDIR /app

COPY --from=builder /go/bin/vego /go/bin/vego

CMD ["/go/bin/vego"]
