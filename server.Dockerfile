# FROM node:16-alpine AS react
# AWS CodeBuild fails due to Docker's pull rate limit, using ECR.
FROM public.ecr.aws/bitnami/node:16 AS react
WORKDIR /usr/src/app
COPY web/react/package*.json ./
RUN npm install
COPY web/react ./
RUN npm run build

# FROM node:16-alpine AS bootstrap
FROM public.ecr.aws/bitnami/node:16 AS bootstrap
WORKDIR /usr/src/app
COPY web/bootstrap/package*.json ./
RUN npm install
COPY cmd/server/template ./ref/
COPY --from=react /usr/src/app/build ./ref/
COPY web/bootstrap ./
RUN npm run css

# Build container for server
# FROM golang:1.17-alpine AS server
FROM public.ecr.aws/bitnami/golang:1.17 AS server
WORKDIR /go/src/app
COPY go.* ./
COPY internal ./internal
RUN go mod download
COPY --from=react /usr/src/app/build ./static/build/
COPY --from=bootstrap /usr/src/app/build ./static/build/
COPY cmd/server ./
# Flag info https://golang.org/cmd/link/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-s' -o server .

FROM scratch
COPY --from=server /go/src/app/server /server
EXPOSE 8080
CMD ["/server"]
