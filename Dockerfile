FROM golang:1.22.7 AS build
WORKDIR /go/src/app
COPY . .
ENV CGO_ENABLED=0 GOOS=linux GOPROXY=direct
RUN go build -v -o app .

FROM alpine:3.20.3
COPY --from=build /go/src/app/app /go/bin/app
EXPOSE 8080
ENTRYPOINT ["/go/bin/app"]