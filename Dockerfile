FROM golang:1.14.2

RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go generate
RUN go build -o main .
CMD ["./main"]
