FROM golang as builder
WORKDIR /build
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o go-shorturl
RUN chmod +x go-shorturl

FROM scratch
WORKDIR /run
COPY --from=builder /build/go-shorturl .
COPY static static
CMD ["./go-shorturl"]