FROM golang:1.22

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN apt update

RUN go build -o /bin/app .

CMD ["/bin/app"]