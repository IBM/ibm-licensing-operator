{{- define "ibm-licensing.watchNamespaces" -}}
{{- $watchNamespaces := list -}}
{{- range (splitList "," (.Values.ibmLicensing.watchNamespace | default "")) -}}
  {{- $watchNamespaces = append $watchNamespaces (trim .) -}}
{{- end -}}
{{- range (splitList "," (.Values.ibmLicensing.spec.watchedNamespaces | default "")) -}}
  {{- $watchNamespaces = append $watchNamespaces (trim .) -}}
{{- end -}}
{{- $watchNamespaces | compact | uniq | join ", " -}}
{{- end -}}