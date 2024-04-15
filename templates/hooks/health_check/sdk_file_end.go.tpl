{{ $CRD := .CRD }}
{{ $SDKAPI := .SDKAPI }}

{{/* Generate helper methods for HealthCheck */}}
{{- range $specFieldName, $specField := $CRD.Config.Resources.HealthCheck.Fields }}
{{- if $specField.From }}
{{- $operationName := $specField.From.Operation }}
{{- $operation := (index $SDKAPI.API.Operations $operationName) -}}
{{- range $healthCheckRefName, $HealthCheckMemberRefs := $operation.InputRef.Shape.MemberRefs -}}
{{- if eq $healthCheckRefName "AddTags" }}
{{- $healthCheckRef := $HealthCheckMemberRefs.Shape.MemberRef }}
{{- $healthCheckRefName = "Tag" }}
func (rm *resourceManager) new{{ $healthCheckRefName }}(
	    c svcapitypes.{{ $healthCheckRefName }},
) *svcsdk.{{ $healthCheckRefName }} {
	res := &svcsdk.{{ $healthCheckRefName }}{}
{{ GoCodeSetSDKForStruct $CRD "" "res" $healthCheckRef "" "c" 1 }}
	return res
}
{{- end }}
{{- end }}
{{- end }}
{{- end }}