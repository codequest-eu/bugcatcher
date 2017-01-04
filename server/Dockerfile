FROM golang:1.7.4

RUN go get github.com/julienschmidt/httprouter
ENV APP_HOME=$GOPATH/src/github.com/codequest-eu/bugcatcher/server
ADD . $APP_HOME
WORKDIR $APP_HOME
RUN go build
EXPOSE 1984
CMD ["./server"]
