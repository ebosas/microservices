FROM golang:1.16-alpine AS backend
WORKDIR /go/src/app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend .
# Info about flags: https://golang.org/cmd/link/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-s' -o backend .

FROM scratch
COPY --from=backend /go/src/app/backend /backend
EXPOSE 8080
CMD ["/backend"]
