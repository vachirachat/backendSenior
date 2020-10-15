FROM golang:1.12-alpine

WORKDIR /api

COPY . .

RUN apk add git
RUN ["go", "get", "./..."]

#RUN ["go", "get", "github.com/githubnemo/CompileDaemon"]
#ENTRYPOINT CompileDaemon -log-prefix=false -build="go build ./cmd/api/" -command="./api"

RUN go build -o api ./cmd/api/
CMD ["./api"]
