module github.com/alexnjh/epsilon/coordinator

go 1.14

require (
	github.com/alexnjh/epsilon/communication v0.0.0-20210129113719-0f00beb417e0
	github.com/alexnjh/epsilon/experiment v0.0.0-20210129113719-0f00beb417e0
	github.com/alexnjh/epsilon/general_purpose_scheduler v0.0.0-20210129113719-0f00beb417e0
	github.com/bigkevmcd/go-configparser v0.0.0-20210106142102-909504547ead
	github.com/prometheus/client_golang v1.9.0
	github.com/sirupsen/logrus v1.7.0
	k8s.io/api v0.20.0
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v0.18.6
)

replace (
	k8s.io/api v0.20.0 => k8s.io/api v0.18.6
	k8s.io/apimachinery v0.20.0 => k8s.io/apimachinery v0.18.6

)
