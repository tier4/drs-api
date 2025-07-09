module github.com/drs-api/tools/client

go 1.22

toolchain go1.24.4

require (
	github.com/drs-api/services/module-agent v0.0.0
	google.golang.org/grpc v1.69.0
)

require (
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/drs-api/services/module-agent => ../../services/module-agent
