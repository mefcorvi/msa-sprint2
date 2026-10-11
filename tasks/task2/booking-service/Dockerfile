FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go test ./... && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/booking-service ./cmd/booking-service

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/booking-service /booking-service
EXPOSE 9090 8080
ENTRYPOINT ["/booking-service"]
