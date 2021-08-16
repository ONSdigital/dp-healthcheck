module github.com/ONSdigital/dp-healthcheck

go 1.15

replace github.com/coreos/etcd v3.3.13+incompatible => github.com/etcd-io/etcd v3.3.25+incompatible

replace github.com/gogo/protobuf v1.2.1 => github.com/gogo/protobuf v1.3.2

require (
	github.com/ONSdigital/dp-api-clients-go v1.42.0 // indirect
	github.com/ONSdigital/log.go/v2 v2.0.5
	github.com/google/go-cmp v0.5.5
	github.com/gopherjs/gopherjs v0.0.0-20210803090616-8f023c250c89 // indirect
	github.com/smartystreets/goconvey v1.6.4
	golang.org/x/sys v0.0.0-20210816032535-30e4713e60e3 // indirect
)
