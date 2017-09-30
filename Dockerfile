# build stage
FROM golang:1.9.0-alpine AS build-env
RUN apk add --update --no-cache git

WORKDIR /go/src/app
RUN go get gopkg.in/src-d/go-git.v4
RUN go get gopkg.in/go-playground/webhooks.v3
RUN go get github.com/Sirupsen/logrus

COPY . .
RUN go build -o reposync

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /go/src/app/reposync /app/
ENTRYPOINT ./reposync
