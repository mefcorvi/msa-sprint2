FROM golang:1.26-alpine AS source
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM source AS test
RUN go test ./...

FROM source AS build
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/booking-service ./cmd/booking-service

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/booking-service /booking-service
EXPOSE 8080 9090
ENTRYPOINT ["/booking-service"]
