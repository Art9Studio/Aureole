module aureole

go 1.16

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
	go.uber.org/atomic => github.com/uber-go/atomic v1.9.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)
