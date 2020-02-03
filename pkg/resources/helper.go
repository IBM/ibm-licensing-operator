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
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var TrueVar = true
var FalseVar = false
var cpu200m = resource.NewMilliQuantity(200, resource.DecimalSI)         // 200m
var memory256Mi = resource.NewQuantity(256*1024*1024, resource.BinarySI) // 256Mi

const APISecretTokenVolumeName = "api-token"
const MeteringAPICertsVolumeName = "metering-api-certs"
const LicensingHTTPSCertsVolumeName = "licensing-https-certs"
const LicensingResourceBase = "ibm-licensing-service"
const LicensingComponentName = "ibmlicensingsvc"
const LicensingReleaseName = "ibmlicensing"

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

var licensingContainerPort int32 = 8080
var licensingTargetPort intstr.IntOrString = intstr.FromInt(8080)

func LabelsForLicensingSelector(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	return map[string]string{"app": GetResourceName(instance), "component": LicensingComponentName, "licensing_cr": instance.GetName()}
}

func LabelsForLicensingMeta(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	return map[string]string{"app.kubernetes.io/name": GetResourceName(instance),
		"app.kubernetes.io/component": LicensingComponentName, "release": LicensingReleaseName}
}

func LabelsForLicensingPod(instance *operatorv1alpha1.IBMLicensing) map[string]string {
	return map[string]string{"app": GetResourceName(instance), "component": LicensingComponentName, "licensing_cr": instance.GetName(),
		"app.kubernetes.io/name": GetResourceName(instance), "app.kubernetes.io/component": LicensingComponentName, "release": LicensingReleaseName}
}
