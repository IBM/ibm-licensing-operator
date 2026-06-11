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

{{/*
Conditional RBAC feature gates (ILS-2352).

Each helper returns the literal string "true" or "false". Gate templates on
  {{- if eq (include "<helper>" .) "true" }}

All gates default to "enabled/present" so that an unset value reproduces today's
RBAC exactly. The nil-safe parenthesised access yields "true" when spec /
features / the field is absent (nil | toString -> "", ne "false" -> true),
mirroring the operator helpers IsNodeCpuCappingEnabled / IsKubeRBACAuthEnabled /
IsOperandRequestsEnabled and IsNamespaceScopeEnabled (nil-default-false: namespace
discovery stays on unless features.nssEnabled is explicitly true).
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
