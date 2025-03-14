{{ $CRD := .CRD }}
{{ $SDKAPI := .SDKAPI }}

{{/* Generate helper methods for HostedZone */}}
{{- range $specFieldName, $specField := $CRD.Config.Resources.HostedZone.Fields }}
{{- if $specField.From }}
{{- $operationName := $specField.From.Operation }}
{{- $operation := (index $SDKAPI.API.Operations $operationName) -}}
{{- range $hostedZoneRefName, $hostedZoneMemberRefs := $operation.InputRef.Shape.MemberRefs -}}
{{- if eq $hostedZoneRefName "AddTags" }}
{{- $hostedZoneRef := $hostedZoneMemberRefs.Shape.MemberRef }}
{{- $hostedZoneRefName = "Tag" }}
func (rm *resourceManager) new{{ $hostedZoneRefName }}(
	    c svcapitypes.{{ $hostedZoneRefName }},
) svcsdktypes.{{ $hostedZoneRefName }} {
	res := svcsdktypes.{{ $hostedZoneRefName }}{}
{{ GoCodeSetSDKForStruct $CRD "" "res" $hostedZoneRef "" "c" 1 }}
	return res
}
{{- end }}
{{- end }}
{{- end }}
{{- end }}

// FromRoute53Tags converts the tags parameter into []*svcapitypes.Tag shape.
func FromRoute53Tags(tags []svcsdktypes.Tag) []*svcapitypes.Tag {
	result := []*svcapitypes.Tag{}
	for _, tag := range tags {
		kCopy := *tag.Key
		vCopy := *tag.Value
		svcapiTag := svcapitypes.Tag{Key: &kCopy, Value: &vCopy}
		result = append(result, &svcapiTag)
	}
	return result
}