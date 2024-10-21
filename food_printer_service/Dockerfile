FROM golang:1.22.2

COPY . /go

RUN go build .

EXPOSE 8181
ENTRYPOINT ["/go/food_printer_service"]
