{{- define "cleanup" }}
{{- $mainBranchName := default "main" .mainBranchName}}
keepImagesBuiltWithinLastNHours: 1
keepPolicies:
  - references:
      branch: /.*/
      limit:
        last: 20
        in: 168h
        operator: And
    imagesPerReference:
      last: 1
      in: 168h
      operator: And
  - references:
      branch: "{{ $mainBranchName }}"
    imagesPerReference:
      last: 1
{{- end }}
