FROM public.ecr.aws/docker/library/golang:tip-alpine3.23 AS builder

RUN apk update && apk --no-cache add gcc git make musl-dev

WORKDIR /go/src

COPY go.mod go.sum ./

RUN go mod tidy && \
    go mod verify

COPY . .

RUN GOOS=linux GOARCH=amd64 make build

FROM public.ecr.aws/docker/library/alpine:3.23

COPY --from=builder go/src/bin/gowatch .

CMD ["./gowatch"]

