FROM golang:alpine as builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY *.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o /storage

FROM busybox
COPY .env /home
COPY --from=builder /storage /home
EXPOSE 80
WORKDIR /home
ENTRYPOINT [ "./storage" ]