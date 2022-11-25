FROM golang:1.19 AS build
ENV APP_HOME /go/app

WORKDIR "$APP_HOME"
COPY . .

RUN go mod download

RUN go build -o devto-mongodb-hackathon github.com/tavsec/devto-mongodb-hackathon-api


FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /go/app/devto-mongodb-hackathon /devto-mongodb-hackathon

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/devto-mongodb-hackathon"]
