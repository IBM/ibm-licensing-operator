{{- define "ibm-licensing.operandRequestsEnabled" -}}
{{- ne ((.Values.ibmLicensing.operandRequests).enabled | toString) "false" -}}
{{- end -}}
