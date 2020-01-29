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
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var TrueVar = true
var FalseVar = false
var cpu200m = resource.NewMilliQuantity(100, resource.DecimalSI)         // 200m
var memory256Mi = resource.NewQuantity(256*1024*1024, resource.BinarySI) // 256Mi

const APISecretTokenVolumeName = "api-token"
const MeteringAPICertsVolumeName = "metering-api-certs"
const LicensingHTTPSCertsVolumeName = "licensing-https-certs"
const LicensingResourceName = "licensing-service"

var licensingContainerPort int32 = 8080
var licensingTargetPort intstr.IntOrString = intstr.FromInt(8080)

func LabelsForLicensingSelector(instanceName string, deploymentName string) map[string]string {
	return map[string]string{"app": deploymentName, "component": "ibmlicensingsvc", "licensing_cr": instanceName}
}

func LabelsForLicensingMeta(deploymentName string) map[string]string {
	return map[string]string{"app.kubernetes.io/name": deploymentName, "app.kubernetes.io/component": "ibmlicensingsvc", "release": "licensing"}
}

func LabelsForLicensingPod(instanceName string, deploymentName string) map[string]string {
	return map[string]string{"app": deploymentName, "component": "ibmlicensingsvc", "licensing_cr": instanceName,
		"app.kubernetes.io/name": deploymentName, "app.kubernetes.io/component": "ibmlicensingsvc", "release": "licensing"}
}
