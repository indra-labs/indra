#!/bin/bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26.0
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.5.0
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.5.0
  echo "Generating root gRPC server protos"

  PROTOS="lightning.proto walletunlocker.proto stateservice.proto **/*.proto"

  # For each of the sub-servers, we then generate their protos, but a restricted
  # set as they don't yet require REST proxies, or swagger docs.
  for file in $PROTOS; do
    DIRECTORY=$(dirname "${file}")
    echo "Generating protos from ${file}, into ${DIRECTORY}"

    # Generate the protos.
    protoc -I/usr/local/include -I. \
      --go_out . --go_opt paths=source_relative \
      --go-grpc_out . --go-grpc_opt paths=source_relative \
      "${file}"

    # Generate the REST reverse proxy.
    annotationsFile=${file//proto/yaml}
    protoc -I/usr/local/include -I. \
      --grpc-gateway_out . \
      --grpc-gateway_opt logtostderr=true \
      --grpc-gateway_opt paths=source_relative \
      --grpc-gateway_opt grpc_api_configuration=${annotationsFile} \
      "${file}"

    # Generate the swagger file which describes the REST API in detail.
    protoc -I/usr/local/include -I. \
      --openapiv2_out . \
      --openapiv2_opt logtostderr=true \
      --openapiv2_opt grpc_api_configuration=${annotationsFile} \
      --openapiv2_opt json_names_for_fields=false \
      "${file}"
done
