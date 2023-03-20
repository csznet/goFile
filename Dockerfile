FROM golang:1.19
WORKDIR /goFile
COPY ./templates /goFile
RUN go get -d -v ./...
RUN go install -v ./...
EXPOSE 8089
CMD ["goFile"]
