FROM golang:1.15-alpine AS build
WORKDIR /src
COPY . .
ENV CGO_ENABLED=0
RUN go mod download
RUN go build -o /out/hello-server server.go

FROM scratch AS bin
COPY --from=build /out/hello-server .
CMD ["./hello-server"]