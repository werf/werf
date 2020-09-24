package werf_chart

var TemplateHelpers = `{{- define "_werf_image" -}}
{{-   $context := index . 0 -}}
{{-   if $context.Values.global.werf.is_stub -}}
{{      $context.Values.global.werf.stub_image }}
{{-   else -}}
{{-     if not $context.Values.global.werf.is_nameless_image -}}
{{-       required "No image specified for template" nil -}}
{{-     end -}}
{{      $context.Values.global.werf.image.docker_image }}
{{-   end -}}
{{- end -}}

{{- define "_werf_image2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   if $context.Values.global.werf.is_stub -}}
{{      $context.Values.global.werf.stub_image }}
{{-   else -}}
{{-     if $context.Values.global.werf.is_nameless_image -}}
{{-       required (printf "No image should be specified for template, got '%s'" $name) nil -}}
{{-     end -}}
{{      index (required (printf "Unknown image '%s' specified for template" $name) (pluck $name $context.Values.global.werf.image | first)) "docker_image" }}
{{-   end -}}
{{- end -}}

{{- define "werf_image" -}}
{{-   if eq (typeOf .) "chartutil.Values" -}}
{{-     $context := . -}}
{{      tuple $context | include "_werf_image" }}
{{-   else if (ge (len .) 2) -}}
{{-     $name := index . 0 -}}
{{-     $context := index . 1 -}}
{{      tuple $name $context | include "_werf_image2" }}
{{-   else -}}
{{-     $context := index . 0 -}}
{{      tuple $context | include "_werf_image" }}
{{-   end -}}
{{- end -}}
`
