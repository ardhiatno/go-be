# --- build stage ---
FROM golang:1.24 AS build

WORKDIR /app
COPY main.go .

RUN go mod init go-be || true
#RUN go mod tidy
RUN go get github.com/valyala/fasthttp github.com/fasthttp/router && \
    go get golang.org/x/image/font && \
    go get golang.org/x/image/font/basicfont && \
    go get golang.org/x/image/math/fixed

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/go-be .

# --- runtime stage (distroless-alpine minimal) ---
FROM alpine:3.20
RUN adduser -D -u 10001 app
USER app
WORKDIR /app
COPY --from=build /out/go-be /app/go-be
EXPOSE 5000
ENV GOMAXPROCS=0
ENTRYPOINT ["/app/go-be","-addr",":5000"]
