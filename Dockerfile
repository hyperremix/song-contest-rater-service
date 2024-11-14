FROM golang:1.21.5 AS build
WORKDIR /go/src/app
COPY . .
ENV CGO_ENABLED=0 GOOS=linux GOPROXY=direct
RUN go build -v -o app .

FROM scratch
COPY --from=build /go/src/app/app /go/bin/app
ENTRYPOINT ["/go/bin/app"]