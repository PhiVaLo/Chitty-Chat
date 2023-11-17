# Additional Commands and notes

## Generate proto files:
```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/proto.proto
```

## Include in git commit message 

```bash
Co-authored-by: Phi Va Lo <phiy@itu.dk>
Co-authored-by: Tien Cam Ly <tily@itu.dk>
Co-authored-by: Patrick Søborg <ptso@itu.dk>
Co-authored-by: Anna Høybye Johansen <annaj@itu.dk>
```