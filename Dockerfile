FROM golang:1.16 as builder

WORKDIR /go/src

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY . .

#Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-s' -o main .

FROM golang:1.16
WORKDIR /go/src
COPY --from=builder /go/src .

RUN chmod 777 main
# RUN chmod 777 start.sh
ENTRYPOINT ["./main"]
