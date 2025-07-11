protoc \
  --proto_path=. \
  --go_out=gen \
  --go_opt=paths=source_relative \ # important
  delivery_v2.proto