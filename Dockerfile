FROM golang:alpine AS build-env

WORKDIR /go/src/github.com/pankrator/payment

COPY go.mod .
COPY go.sum .
RUN go mod download


COPY . .
# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /payment

FROM scratch

# RUN apk add --update --no-cache ca-certificates git

COPY --from=build-env /go/src/github.com/pankrator/payment/config.yaml .
COPY --from=build-env /go/src/github.com/pankrator/payment/users.csv .
COPY --from=build-env /go/src/github.com/pankrator/payment/templates ./templates
COPY --from=build-env /go/src/github.com/pankrator/payment/storage/gormdb/migrations /go/src/github.com/pankrator/payment/storage/gormdb/migrations
COPY --from=build-env /payment /payment

EXPOSE 8000

ENTRYPOINT ["/payment"]