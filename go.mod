module github.com/ibm/ibm-licensing-operator

go 1.15

require (
	github.com/coreos/prometheus-operator v0.41.0
	github.com/go-logr/logr v0.3.0
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/openshift/api v0.0.0-20200930075302-db52bc4ef99f
	github.com/redhat-marketplace/redhat-marketplace-operator/v2 v2.0.0-20210125205956-4eda6b4abf4e
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.4
)

replace k8s.io/client-go => k8s.io/client-go v0.19.4

replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
