FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /procwatch ./cmd/procwatch

FROM alpine:3.20
RUN adduser -D app
USER app
COPY --from=build /procwatch /procwatch
EXPOSE 8080
CMD ["/procwatch", "serve"]
