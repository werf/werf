package helpers

var ChartTemplateHelpers = `{{- define "_werf_image" -}}
{{-   $context := index . 0 -}}
{{-   if $context.Values.werf.is_stub -}}
{{      $context.Values.werf.stub_image }}
{{-   else -}}
{{-     if not $context.Values.werf.is_nameless_image -}}
{{-       required "No image specified for template" nil -}}
{{-     end -}}
{{      $context.Values.werf.nameless_image }}
{{-   end -}}
{{- end -}}

{{- define "_werf_image2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   if $context.Values.werf.is_stub -}}
{{      $context.Values.werf.stub_image }}
{{-   else -}}
{{-     if $context.Values.werf.is_nameless_image -}}
{{-       required (printf "No image should be specified for template, got '%s'" $name) nil -}}
{{-     end -}}
{{      required (printf "Unknown image '%s' specified for template" $name) (pluck $name $context.Values.werf.image | first) }}
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
