FROM golang:1.19.2

COPY ./ /

WORKDIR /cmd

CMD go build main.go && ./main