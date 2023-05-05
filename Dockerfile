FROM golang

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY main.go .
RUN go build -o app

CMD [ "./app" ]
