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

{{- define "ibm-licensing.nodeCpuCappingEnabled" -}}
{{- ne (((.Values.ibmLicensing.spec).features).nodeCpuCappingEnabled | toString) "false" -}}
{{- end -}}

{{- define "ibm-licensing.kubeRBACAuthEnabled" -}}
{{- ne (((.Values.ibmLicensing.spec).features).kubeRBACAuthEnabled | toString) "false" -}}
{{- end -}}

{{- define "ibm-licensing.operandRequestsEnabled" -}}
{{- ne (((.Values.ibmLicensing.spec).features).operandRequestsEnabled | toString) "false" -}}
{{- end -}}

{{- define "ibm-licensing.namespaceDiscoveryEnabled" -}}
{{- ne (((.Values.ibmLicensing.spec).features).nssEnabled | toString) "true" -}}
{{- end -}}

{{/* True when the unrestricted operand SA/ClusterRole is the active one (nss off). */}}
{{- define "ibm-licensing.unrestrictedClusterRoleNeeded" -}}
{{- ne (((.Values.ibmLicensing.spec).features).nssEnabled | toString) "true" -}}
{{- end -}}

{{/* Restricted ClusterRole is the operand role only in nss mode (it backs the
     ibm-license-service-restricted SA). Outside nss mode the unrestricted role
     already carries the node/kube-RBAC rules, so the restricted object would be
     dead RBAC. Render it only when nss is on AND a cluster-scoped rule remains:
     nodes (capping) or tokenreviews/SAR (kubeRBACAuth). */}}
{{- define "ibm-licensing.restrictedClusterRoleNeeded" -}}
{{- and (eq (include "ibm-licensing.namespaceDiscoveryEnabled" .) "false") (or (eq (include "ibm-licensing.nodeCpuCappingEnabled" .) "true") (eq (include "ibm-licensing.kubeRBACAuthEnabled" .) "true")) -}}
{{- end -}}

{{/* cluster-monitoring-view binding: only for datasource=prometheus. */}}
{{- define "ibm-licensing.clusterMonitoringNeeded" -}}
{{- eq ((.Values.ibmLicensing.spec).datasource | toString) "prometheus" -}}
{{- end -}}

{{/* The operand ServiceAccount in use: restricted when nss is on, default otherwise.
     Drives the cluster-monitoring-view binding subject so it follows the active SA. */}}
{{- define "ibm-licensing.operandServiceAccount" -}}
{{- if eq (include "ibm-licensing.namespaceDiscoveryEnabled" .) "false" -}}
ibm-license-service-restricted
{{- else -}}
ibm-license-service
{{- end -}}
{{- end -}}
