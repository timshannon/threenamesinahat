FROM golang:1.22.1

RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go generate
RUN go build -o main .
CMD ["./main"]
