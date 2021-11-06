# FROM golang:1.17-alpine AS database
# AWS CodeBuild fails due to Docker's pull rate limit, using ECR.
FROM public.ecr.aws/bitnami/golang:1.17 AS database
WORKDIR /go/src/app
COPY go.* ./
COPY internal ./internal
RUN go mod download
COPY cmd/database ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-s' -o database .

FROM scratch
COPY --from=database /go/src/app/database /database
CMD ["/database"]
