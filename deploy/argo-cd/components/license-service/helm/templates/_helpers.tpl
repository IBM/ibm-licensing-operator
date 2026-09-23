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

{{/* Restricted ClusterRole is the operand role only in nss mode (it backs the
     ibm-license-service-restricted SA). Outside nss mode the unrestricted role
     is the active one, so the restricted object would be dead RBAC. Within nss
     mode every rule block it carries is feature-gated (nodes via CPU capping,
     the metadata/definition/querysource CR reads via customResources, and the
     kube-RBAC auth rules), so render it only when at least one of those is on;
     otherwise it would be an empty ClusterRole plus a dangling binding. */}}
{{- define "ibm-licensing.restrictedClusterRoleNeeded" -}}
{{- and ((.Values.ibmLicensing.spec).features).nssEnabled (or ((.Values.ibmLicensing.spec).features).nodeCpuCappingEnabled ((.Values.ibmLicensing.spec).features).customResourcesEnabled ((.Values.ibmLicensing.spec).features).kubeRBACAuthEnabled) -}}
{{- end -}}

{{/* The operand ServiceAccount in use: restricted when nss is on, default otherwise. */}}
{{- define "ibm-licensing.operandServiceAccount" -}}
{{- if ((.Values.ibmLicensing.spec).features).nssEnabled -}}
ibm-license-service-restricted
{{- else -}}
ibm-license-service
{{- end -}}
{{- end -}}
