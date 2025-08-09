## Multi-stage build for go-sixtysix example HTTP server
## Usage:
##   docker build -t go-sixtysix:dev .
##   docker run -p 8080:8080 go-sixtysix:dev

FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod .
RUN go mod download
COPY . .
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" -o /out/server ./examples/server

FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=build /out/server /app/server
EXPOSE 8080
USER nonroot:nonroot
ENV PORT=8080
ENTRYPOINT ["/app/server"]
