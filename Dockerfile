FROM golang:alpine as builder

RUN mkdir /build

ADD . /build/

WORKDIR /build

# Statically link runtime libraries into binary
# Else "standard_init_linux.go:228: exec user process caused: no such file or directory"
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main ./src

FROM scratch

COPY --from=builder /build/main /app/
WORKDIR /app
CMD ["./main"]
EXPOSE 8080
