FROM golang:1.14

WORKDIR /go/src/app

ENV JWT_KEY="your-jwt-key"

COPY . .

RUN go mod download && go install ./... && go build -o main .

EXPOSE 8080

CMD ["./main"]