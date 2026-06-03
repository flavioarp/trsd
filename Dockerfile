FROM golang:1.22

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN apt update && apt install -y ghostscript

RUN go build -o /bin/app .

CMD ["/bin/app"]