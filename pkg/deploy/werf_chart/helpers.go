package werf_chart

var WerfChartHelpersTpl = []byte(`{{- define "werf_secret_file" -}}
{{-   $relative_file_path := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   $context.Files.Get (print "` + WerfChartDecodedSecretDir + `/" $relative_file_path) -}}
{{- end -}}

{{- define "_image" -}}
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

{{- define "_werf_container__imagePullPolicy" -}}
{{-   $context := index . 0 -}}
{{-   if $context.Values.global.werf.ci.is_branch -}}
imagePullPolicy: Always
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
{{      tuple $context | include "_werf_container__imagePullPolicy" }}
{{-   else if (ge (len .) 2) -}}
{{-     $name := index . 0 -}}
{{-     $context := index . 1 -}}
{{      tuple $name $context | include "_werf_container__image2" }}
{{      tuple $context | include "_werf_container__imagePullPolicy" }}
{{-   else -}}
{{-     $context := index . 0 -}}
{{      tuple $context | include "_werf_container__image" }}
{{      tuple $context | include "_werf_container__imagePullPolicy" }}
{{-   end -}}
{{- end -}}

{{- define "_image_id" -}}
{{-   $context := index . 0 -}}
{{-   if not $context.Values.global.werf.is_nameless_image -}}
{{-     required "No image specified for template" nil -}}
{{-   end -}}
{{    $context.Values.global.werf.image.docker_image_id }}
{{- end -}}

{{- define "_image_id2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   if $context.Values.global.werf.is_nameless_image -}}
{{-     required (printf "No image should be specified for template, got '%s'" $name) nil -}}
{{-   end -}}
{{    index (required (printf "Unknown image '%s' specified for template" $name) (pluck $name $context.Values.global.werf.image | first)) "docker_image_id" }}
{{- end -}}

{{- define "image_id" -}}
{{-   if eq (typeOf .) "chartutil.Values" -}}
{{-     $context := . -}}
{{      tuple $context | include "_image_id" }}
{{-   else if (ge (len .) 2) -}}
{{-     $name := index . 0 -}}
{{-     $context := index . 1 -}}
{{      tuple $name $context | include "_image_id2" }}
{{-   else -}}
{{-     $context := index . 0 -}}
{{      tuple $context | include "_image_id" }}
{{-   end -}}
{{- end -}}

{{- define "_werf_container_env" -}}
{{-   $context := index . 0 -}}
{{-   if $context.Values.global.werf.ci.is_branch -}}
- name: DOCKER_IMAGE_ID
  value: {{ tuple $context | include "_image_id" }}
{{-   else -}}
- name: DOCKER_IMAGE_ID
  value: "-"
{{-   end -}}
{{- end -}}

{{- define "_werf_container_env2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   if $context.Values.global.werf.ci.is_branch -}}
- name: DOCKER_IMAGE_ID
  value: {{ tuple $name $context | include "_image_id2" }}
{{-   else -}}
- name: DOCKER_IMAGE_ID
  value: "-"
{{-   end -}}
{{- end -}}

{{- define "werf_container_env" -}}
{{-   if eq (typeOf .) "chartutil.Values" -}}
{{-     $context := . -}}
{{      tuple $context | include "_werf_container_env" }}
{{-   else if (ge (len .) 2) -}}
{{-     $name := index . 0 -}}
{{-     $context := index . 1 -}}
{{      tuple $name $context | include "_werf_container_env2" }}
{{-   else -}}
{{-     $context := index . 0 -}}
{{      tuple $context | include "_werf_container_env" }}
{{-   end -}}
{{- end -}}
`)
