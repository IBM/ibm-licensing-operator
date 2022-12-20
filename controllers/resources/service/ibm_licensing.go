package service

import (
	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetDefaultIBMLicensing() operatorv1alpha1.IBMLicensing {
	return operatorv1alpha1.IBMLicensing{
		ObjectMeta: metav1.ObjectMeta{
			Name: "instance",
		},
		Spec: operatorv1alpha1.IBMLicensingSpec{
			Datasource:  "datacollector",
			HTTPSEnable: true,
		},
	}
}
