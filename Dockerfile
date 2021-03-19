FROM golang:1.16.2-alpine3.13

WORKDIR /app/omnikanji

COPY . .

RUN go get -v -d ./...
RUN go install -v ./...


CMD ["omnikanji"]

