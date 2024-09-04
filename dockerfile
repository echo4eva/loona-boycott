FROM golang:1.22.5

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o bin ./cmd/web

EXPOSE 8080

ENTRYPOINT [ "/app/bin" ]