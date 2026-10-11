GO_IMAGE := golang:1.26-alpine
PROTOC_GEN_GO := google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
PROTOC_GEN_GO_GRPC := google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
GO_PACKAGE := hotelio/booking-service/internal/pb

.PHONY: proto test

# Generates internal/pb from the shared tasks/task2/booking.proto (protoc runs in a container).
proto:
	docker run --rm -v $(CURDIR)/..:/work -w /work $(GO_IMAGE) sh -c '\
		apk add --no-cache protobuf protobuf-dev >/dev/null && \
		go install $(PROTOC_GEN_GO) && go install $(PROTOC_GEN_GO_GRPC) && \
		protoc -I . \
			--go_out=booking-service/internal/pb --go_opt=paths=source_relative --go_opt=Mbooking.proto=$(GO_PACKAGE) \
			--go-grpc_out=booking-service/internal/pb --go-grpc_opt=paths=source_relative --go-grpc_opt=Mbooking.proto=$(GO_PACKAGE) \
			booking.proto'

test:
	go test ./...
