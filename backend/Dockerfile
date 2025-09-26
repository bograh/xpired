FROM golang:1.25-alpine
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping ./cmd/server/main.go

EXPOSE 8080

CMD ["/docker-gs-ping"]