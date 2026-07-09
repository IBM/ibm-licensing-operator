{{/*
Copyright 2026 IBM Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/}}

{{/* True when the restricted ClusterRole would carry at least one rule. Render the ClusterRole
     and its binding only then; otherwise they would be an empty ClusterRole plus a dangling binding.
     Covers: nodeCpuCapping, customResources, kubeRBACAuth, and chargebackEnabled. */}}
{{- define "ibm-licensing.restrictedClusterRoleNotEmpty" -}}
{{- or ((.Values.ibmLicensing.spec).features).nodeCpuCappingEnabled ((.Values.ibmLicensing.spec).features).customResourcesEnabled ((.Values.ibmLicensing.spec).features).kubeRBACAuthEnabled .Values.ibmLicensing.spec.chargebackEnabled -}}
{{- end -}}

{{/* The operand ServiceAccount in use: restricted when nss is on, default otherwise.
     Drives the cluster-monitoring-view binding subject so it follows the active SA. */}}
{{- define "ibm-licensing.operandServiceAccount" -}}
{{- if ((.Values.ibmLicensing.spec).features).nssEnabled -}}
ibm-license-service-restricted
{{- else -}}
ibm-license-service
{{- end -}}
{{- end -}}
