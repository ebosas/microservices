FROM node:14-alpine AS react
WORKDIR /usr/src/app
COPY react/package*.json .
RUN npm install
COPY react .
RUN npm run build

FROM node:14-alpine AS bootstrap
WORKDIR /usr/src/app
COPY bootstrap/package*.json .
RUN npm install
COPY server/template.html ./ref/
COPY --from=react /usr/src/app/build ./ref/
COPY bootstrap .
RUN npm run css

# Build container for server
FROM golang:1.16-alpine AS server
WORKDIR /go/src/app
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY --from=react /usr/src/app/build ./static/
COPY --from=bootstrap /usr/src/app/build ./static/
COPY server .
# Info about flags: https://golang.org/cmd/link/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-s' -o server .

FROM scratch
COPY --from=server /go/src/app/server /server
EXPOSE 8080
CMD ["/server"]
