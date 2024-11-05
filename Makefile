.SILENT:

proto-generate-go:
	protoc \
	-I./protos \
	--go_out=./server/pkg/proto \
	--go_opt=paths=source_relative \
    --go-grpc_out=./server/pkg/proto \
	--go-grpc_opt=paths=source_relative \
	./protos/*.proto

proto-generate-py:
	python -m grpc_tools.protoc \
	-I./protos \
	--python_out=./recommender/proto \
	--pyi_out=./recommender/proto \
	--grpc_python_out=./recommender/proto \
	./protos/*.proto

proto-generate: proto-generate-go proto-generate-py