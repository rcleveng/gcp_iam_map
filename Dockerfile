FROM golang:1.22 as build
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o app .

FROM gcr.io/distroless/base
COPY --from=build /go/src/app/. /
COPY --from=build /go/src/app/public/ /public/
CMD ["/app", "server", "--port", "8080"]