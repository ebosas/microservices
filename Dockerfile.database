FROM golang:1.16-alpine AS database
WORKDIR /go/src/app
COPY database/go.mod database/go.sum ./
RUN go mod download
COPY database .
# Info about flags: https://golang.org/cmd/link/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-s' -o database .

FROM scratch
COPY --from=database /go/src/app/database /database
EXPOSE 8080
CMD ["/database"]
