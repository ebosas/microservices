FROM golang:1.16-alpine AS backend
WORKDIR /go/src/app
COPY go.* .
COPY internal ./internal
RUN go mod download
COPY cmd/backend .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-s' -o backend .

FROM scratch
COPY --from=backend /go/src/app/backend /backend
EXPOSE 8080
CMD ["/backend"]
