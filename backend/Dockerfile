FROM golang:1.22-alpine

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod .

RUN go mod tidy

COPY . .

RUN go build -o main .

CMD ["./main"] 