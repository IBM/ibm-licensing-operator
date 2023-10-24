//
// Copyright 2023 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package service

import (
	"time"

	rhmpcommon "github.com/redhat-marketplace/redhat-marketplace-operator/v2/apis/marketplace/common"
	rhmp "github.com/redhat-marketplace/redhat-marketplace-operator/v2/apis/marketplace/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
)

func GetMeterDefinitionList(instance *operatorv1alpha1.IBMLicensing) []*rhmp.MeterDefinition {
	return []*rhmp.MeterDefinition{
		getCloudPakMeterDefinition(instance),
		getProductMeterDefinition(instance),
		getChargebackMeterDefinition(instance),
		getServiceMeterDefinition(instance)}

}

func getCloudPakMeterDefinition(instance *operatorv1alpha1.IBMLicensing) *rhmp.MeterDefinition {
	return &rhmp.MeterDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetMeterDefinitionName(instance, "product"),
			Namespace: instance.Spec.InstanceNamespace,
		},
		Spec: rhmp.MeterDefinitionSpec{
			Group: "{{ .Label.productId}}.licensing.ibm.com",
			Kind:  "IBMLicensing",
			ResourceFilters: []rhmp.ResourceFilter{
				{
					Namespace: &rhmp.NamespaceFilter{
						UseOperatorGroup: true,
					},
					OwnerCRD: &rhmp.OwnerCRDFilter{
						GroupVersionKind: rhmpcommon.GroupVersionKind{
							APIVersion: "operator.ibm.com/v1alpha1",
							Kind:       "IBMLicensing",
						},
					},
					WorkloadType: rhmpcommon.WorkloadTypeService,
				},
			},
			Meters: []rhmp.MeterWorkload{
				{
					Name:               "{{ .Label.productId}}.licensing.ibm.com",
					Aggregation:        "max",
					Period:             &metav1.Duration{Duration: 24 * time.Hour},
					WorkloadType:       rhmpcommon.WorkloadTypeService,
					Metric:             "{{ .Label.metricId}}",
					Query:              "avg_over_time(product_license_usage{}[1d])",
					GroupBy:            []string{"metricId", "productId"},
					ValueLabelOverride: "{{ .Label.value}}",
					DateLabelOverride:  "{{ .Label.date}}",
				},
			},
		},
	}
}

func getProductMeterDefinition(instance *operatorv1alpha1.IBMLicensing) *rhmp.MeterDefinition {
	return &rhmp.MeterDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetMeterDefinitionName(instance, "bundleproduct"),
			Namespace: instance.Spec.InstanceNamespace,
		},
		Spec: rhmp.MeterDefinitionSpec{
			Group: "{{ .Label.productId}}.licensing.ibm.com",
			Kind:  "IBMLicensing-Bundle",
			ResourceFilters: []rhmp.ResourceFilter{
				{
					Namespace: &rhmp.NamespaceFilter{
						UseOperatorGroup: true,
					},
					OwnerCRD: &rhmp.OwnerCRDFilter{
						GroupVersionKind: rhmpcommon.GroupVersionKind{
							APIVersion: "operator.ibm.com/v1alpha1",
							Kind:       "IBMLicensing",
						},
					},
					WorkloadType: rhmpcommon.WorkloadTypeService,
				},
			},
			Meters: []rhmp.MeterWorkload{
				{
					Name:               "{{ .Label.productId}}.licensing.ibm.com",
					Aggregation:        "max",
					Period:             &metav1.Duration{Duration: 24 * time.Hour},
					WorkloadType:       rhmpcommon.WorkloadTypeService,
					Metric:             "{{ .Label.parentMetricId}}",
					Query:              "avg_over_time(product_license_usage_details{}[1d])",
					GroupBy:            []string{"metricId", "productId", "parentMetricId", "parentProductId", "productConversionRatio"},
					ValueLabelOverride: "{{ .Label.value}}",
					DateLabelOverride:  "{{ .Label.date}}",
				},
			},
		},
	}
}

func getServiceMeterDefinition(instance *operatorv1alpha1.IBMLicensing) *rhmp.MeterDefinition {
	return &rhmp.MeterDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetMeterDefinitionName(instance, "service"),
			Namespace: instance.Spec.InstanceNamespace,
		},
		Spec: rhmp.MeterDefinitionSpec{
			Group: "{{ .Label.productId}}.licensing.ibm.com",
			Kind:  "IBMLicensing-Service",
			ResourceFilters: []rhmp.ResourceFilter{
				{
					Namespace: &rhmp.NamespaceFilter{
						UseOperatorGroup: true,
					},
					OwnerCRD: &rhmp.OwnerCRDFilter{
						GroupVersionKind: rhmpcommon.GroupVersionKind{
							APIVersion: "operator.ibm.com/v1alpha1",
							Kind:       "IBMLicensing",
						},
					},
					WorkloadType: rhmpcommon.WorkloadTypeService,
				},
			},
			Meters: []rhmp.MeterWorkload{
				{
					Name:               "Cp4d Capability",
					Aggregation:        "max",
					Period:             &metav1.Duration{Duration: 24 * time.Hour},
					WorkloadType:       rhmpcommon.WorkloadTypeService,
					Metric:             "{{ .Label.parentMetricId}}",
					Query:              "avg_over_time(cp4d_capability{}[1d])",
					GroupBy:            []string{"metricId", "productId", "parentMetricId", "parentProductId", "topLevelProductId", "topLevelMetricId"},
					ValueLabelOverride: "{{ .Label.value}}",
					DateLabelOverride:  "{{ .Label.date}}",
				},
			},
		},
	}
}

func getChargebackMeterDefinition(instance *operatorv1alpha1.IBMLicensing) *rhmp.MeterDefinition {
	return &rhmp.MeterDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetMeterDefinitionName(instance, "chargeback"),
			Namespace: instance.Spec.InstanceNamespace,
		},
		Spec: rhmp.MeterDefinitionSpec{
			Group: "{{ .Label.productId}}.licensing.ibm.com",
			Kind:  "IBMLicensing-{{ .Label.groupName}}",
			ResourceFilters: []rhmp.ResourceFilter{
				{
					Namespace: &rhmp.NamespaceFilter{
						UseOperatorGroup: true,
					},
					OwnerCRD: &rhmp.OwnerCRDFilter{
						GroupVersionKind: rhmpcommon.GroupVersionKind{
							APIVersion: "operator.ibm.com/v1alpha1",
							Kind:       "IBMLicensing",
						},
					},
					WorkloadType: rhmpcommon.WorkloadTypeService,
				},
			},
			Meters: []rhmp.MeterWorkload{
				{
					Name:               "{{ .Label.productId}}.licensing.ibm.com",
					Aggregation:        "max",
					Period:             &metav1.Duration{Duration: 24 * time.Hour},
					WorkloadType:       rhmpcommon.WorkloadTypeService,
					Metric:             "{{ .Label.parentMetricId}}",
					Query:              "avg_over_time(product_license_usage_chargeback{}[1d])",
					GroupBy:            []string{"metricId", "productId", "parentMetricId", "parentProductId", "productConversionRatio"},
					ValueLabelOverride: "{{ .Label.value}}",
					DateLabelOverride:  "{{ .Label.date}}",
				},
			},
		},
	}
}

func GetMeterDefinitionName(instance *operatorv1alpha1.IBMLicensing, meterType string) string {
	return LicensingResourceBase + "-" + meterType + "-" + instance.GetName()

}
