FROM golang:latest
WORKDIR /go/src/goFile
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...
EXPOSE 8089
CMD ["goFile"]
