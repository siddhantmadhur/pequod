FROM golang:1.22.2

COPY . /app
WORKDIR /app


RUN go build -o bin ./cmd/server

EXPOSE 8080

CMD [ "./bin" ]
