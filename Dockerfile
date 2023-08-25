FROM golang:alpine

WORKDIR /app/segment-api

COPY go.mod .
COPY go.sum .
ENV GOPATH=/
RUN go mod download

#build appliction
COPY . .
RUN go build -o segment-api ./cmd/app/main.go

CMD ["./segment-api"]