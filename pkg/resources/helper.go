//
// Copyright 2020 IBM Corporation
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

package resources

import (
	"math/rand"
	"time"

	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// cannot set to const due to k8s struct needing pointers to primitive types

var TrueVar = true
var FalseVar = false
var cpu200m = resource.NewMilliQuantity(200, resource.DecimalSI)
var memory256Mi = resource.NewQuantity(256*1024*1024, resource.BinarySI)
var cpu500m = resource.NewMilliQuantity(500, resource.DecimalSI)
var memory512Mi = resource.NewQuantity(512*1024*1024, resource.BinarySI)

// TODO: validate if good mode, in helm chart was 0644
var defaultSecretMode int32 = 420
var seconds60 int64 = 60

const LicensingResourceBase = "ibm-licensing-service"
const LicensingComponentName = "ibm-licensing-service-svc"
const LicensingReleaseName = "ibm-licensing-service"
const LicensingProductName = "IBM Cloud Platform Common Services"
const LicensingProductID = "1234-567"
const LicensingProductVersion = "1.0.0"

const randStringCharset string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const randStringCharsetLength = len(randStringCharset)

func RandString(length int) string {
	randFunc := rand.New(rand.NewSource(time.Now().UnixNano()))
	outputStringByte := make([]byte, length)
	for i := 0; i < length; i++ {
		outputStringByte[i] = randStringCharset[randFunc.Intn(randStringCharsetLength)]
	}
	return string(outputStringByte)
}

func GetResourceName(instance *operatorv1alpha1.IBMLicensing) string {
	return LicensingResourceBase + "-" + instance.GetName()
}

func LabelsForLicensingSelector(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	return map[string]string{"app": GetResourceName(instance), "component": LicensingComponentName, "licensing_cr": instance.GetName()}
}

func LabelsForLicensingMeta(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	return map[string]string{"app.kubernetes.io/name": GetResourceName(instance), "app.kubernetes.io/component": LicensingComponentName,
		"app.kubernetes.io/managed-by": "operator", "app.kubernetes.io/instance": LicensingReleaseName, "release": LicensingReleaseName}
}

func AnnotationsForPod() map[string]string {
	return map[string]string{"productName": LicensingProductName,
		"productID": LicensingProductID, "productVersion": LicensingProductVersion,
		"clusterhealth.ibm.com/dependencies": "ibm-common-services.metering"}
}

func LabelsForLicensingPod(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	podLabels := LabelsForLicensingMeta(instance)
	selectorLabels := LabelsForLicensingSelector(instance)
	for key, value := range selectorLabels {
		podLabels[key] = value
	}
	return podLabels
}

func Contains(s []corev1.LocalObjectReference, e corev1.LocalObjectReference) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
