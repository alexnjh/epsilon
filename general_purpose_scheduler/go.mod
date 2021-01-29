module github.com/alexnjh/epsilon/general_purpose_scheduler

go 1.14

require (
  github.com/alexnjh/epsilon/communication v0.0.0-20210129103033-c81fe4affdf0
	github.com/MichaelTJones/pcg v0.0.0-20180122055547-df440c6ed7ed
	github.com/bigkevmcd/go-configparser v0.0.0-20200217161103-d137835d2579
	github.com/davidminor/gorand v0.0.0-20161120223607-283446f2caf5
	github.com/davidminor/uint128 v0.0.0-20141227063632-5745f1bf8041 // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/json-iterator/go v1.1.8
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/streadway/amqp v1.0.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/apiserver v0.18.6
	k8s.io/client-go v0.18.6
	k8s.io/component-base v0.18.6
	k8s.io/csi-translation-lib v0.18.6
	k8s.io/klog v1.0.0
	k8s.io/kube-scheduler v0.18.6
	k8s.io/utils v0.0.0-20200619165400-6e3d28b6ed19
	sigs.k8s.io/yaml v1.2.0
)
