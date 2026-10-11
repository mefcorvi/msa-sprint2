FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go test ./... && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/booking-history-service ./cmd/booking-history-service

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/booking-history-service /booking-history-service
EXPOSE 8080
ENTRYPOINT ["/booking-history-service"]
