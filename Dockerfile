FROM golang:1.17-alpine

RUN apk update && apk upgrade && apk add --no-cache bash git && apk add --no-cache chromium

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY /webhook/* ./webhook/
COPY /services/* ./services/
COPY /config/* ./config/
COPY /cronTasks/* ./cronTasks/

RUN go build -o /snpgolang

EXPOSE 8083

CMD [ "/[PROJECTNAME]" ]