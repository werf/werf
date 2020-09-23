package werf_chart

var TemplateHelpers = `{{- define "_image" -}}
{{-   $context := index . 0 -}}
{{-   if not $context.Values.global.werf.is_nameless_image -}}
{{-     required "No image specified for template" nil -}}
{{-   end -}}
{{    $context.Values.global.werf.image.docker_image }}
{{- end -}}

{{- define "_image2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   if $context.Values.global.werf.is_nameless_image -}}
{{-     required (printf "No image should be specified for template, got '%s'" $name) nil -}}
{{-   end -}}
{{    index (required (printf "Unknown image '%s' specified for template" $name) (pluck $name $context.Values.global.werf.image | first)) "docker_image" }}
{{- end -}}

{{- define "image" -}}
{{-   if eq (typeOf .) "chartutil.Values" -}}
{{-     $context := . -}}
{{      tuple $context | include "_image" }}
{{-   else if (ge (len .) 2) -}}
{{-     $name := index . 0 -}}
{{-     $context := index . 1 -}}
{{      tuple $name $context | include "_image2" }}
{{-   else -}}
{{-     $context := index . 0 -}}
{{      tuple $context | include "_image" }}
{{-   end -}}
{{- end -}}

{{- define "_werf_container__image" -}}
{{-   $context := index . 0 -}}
image: {{ tuple $context | include "_image" }}
{{- end -}}

{{- define "_werf_container__image2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
image: {{ tuple $name $context | include "_image2" }}
{{- end -}}

{{- define "werf_container_image" -}}
{{-   if eq (typeOf .) "chartutil.Values" -}}
{{-     $context := . -}}
{{      tuple $context | include "_werf_container__image" }}
{{-   else if (ge (len .) 2) -}}
{{-     $name := index . 0 -}}
{{-     $context := index . 1 -}}
{{      tuple $name $context | include "_werf_container__image2" }}
{{-   else -}}
{{-     $context := index . 0 -}}
{{      tuple $context | include "_werf_container__image" }}
{{-   end -}}
{{- end -}}
`
