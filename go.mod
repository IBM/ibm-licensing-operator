module github.com/ibm/ibm-licensing-operator

go 1.15

require (
	github.com/cloudflare/cfssl v1.5.0
	github.com/coreos/prometheus-operator v0.40.0
	github.com/go-logr/logr v0.1.0
	github.com/openshift/api v0.0.0-20200205133042-34f0ec8dab87
	github.com/operator-framework/operator-sdk v0.19.4
	github.com/redhat-marketplace/redhat-marketplace-operator v0.0.0-20201211175424-6b3ce5b64e99
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.12
	k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery v0.18.12
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	sigs.k8s.io/controller-runtime v0.6.3
)

// Pinned to kubernetes-1.16.2
replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.4 // Required by prometheus-operator
)
