module github.com/alexnjh/epsilon/coordinator

go 1.14

require (
	cloud.google.com/go v0.54.0 // indirect
	github.com/Azure/go-autorest/autorest v0.11.1 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.5 // indirect
	github.com/alexnjh/epsilon v0.0.2 // indirect
	github.com/alexnjh/epsilon/communication v0.0.0-20210325152251-2532cd9a7283
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/bigkevmcd/go-configparser v0.0.0-20210106142102-909504547ead
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible // indirect
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/googleapis/gnostic v0.4.1
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/moby/term v0.0.0-20200312100748-672ec06f55cd // indirect
	github.com/prometheus/client_golang v1.9.0
	github.com/sirupsen/logrus v1.7.0
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/stretchr/testify v1.6.1 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	go.etcd.io/bbolt v1.3.5 // indirect
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489 // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	k8s.io/api v0.20.0
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/gengo v0.0.0-20200413195148-3a45101e95ac // indirect
	k8s.io/klog/v2 v2.4.0 // indirect
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.14 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.0.2 // indirect
)

replace (
	k8s.io/api v0.20.0 => k8s.io/api v0.19.2
	k8s.io/apimachinery v0.20.0 => k8s.io/apimachinery v0.19.2
	k8s.io/client-go v11.0.0+incompatible => k8s.io/client-go v0.19.2
)
