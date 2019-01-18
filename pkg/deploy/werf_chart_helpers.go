package deploy

var WerfChartHelpersTpl = []byte(`{{- define "werf_secret_file" -}}
{{-   $relative_file_path := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   $context.Files.Get (print "` + WerfChartDecodedSecretDir + `/" $relative_file_path) -}}
{{- end -}}

{{- define "_dimg" -}}
{{-   $context := index . 0 -}}
{{-   if not $context.Values.global.werf.is_nameless_dimg -}}
{{-     required "No dimg specified for template" nil -}}
{{-   end -}}
{{    $context.Values.global.werf.dimg.docker_image }}
{{- end -}}

{{- define "_dimg2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   if $context.Values.global.werf.is_nameless_dimg -}}
{{-     required (printf "No dimg should be specified for template, got '%s'" $name) nil -}}
{{-   end -}}
{{    index (required (printf "Unknown dimg '%s' specified for template" $name) (pluck $name $context.Values.global.werf.dimg | first)) "docker_image" }}
{{- end -}}

{{- define "dimg" -}}
{{-   if eq (typeOf .) "chartutil.Values" -}}
{{-     $context := . -}}
{{      tuple $context | include "_dimg" }}
{{-   else if (ge (len .) 2) -}}
{{-     $name := index . 0 -}}
{{-     $context := index . 1 -}}
{{      tuple $name $context | include "_dimg2" }}
{{-   else -}}
{{-     $context := index . 0 -}}
{{      tuple $context | include "_dimg" }}
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
image: {{ tuple $context | include "_dimg" }}
{{- end -}}

{{- define "_werf_container__image2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
image: {{ tuple $name $context | include "_dimg2" }}
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

{{- define "_dimg_id" -}}
{{-   $context := index . 0 -}}
{{-   if not $context.Values.global.werf.is_nameless_dimg -}}
{{-     required "No dimg specified for template" nil -}}
{{-   end -}}
{{    $context.Values.global.werf.dimg.docker_image_id }}
{{- end -}}

{{- define "_dimg_id2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   if $context.Values.global.werf.is_nameless_dimg -}}
{{-     required (printf "No dimg should be specified for template, got '%s'" $name) nil -}}
{{-   end -}}
{{    index (required (printf "Unknown dimg '%s' specified for template" $name) (pluck $name $context.Values.global.werf.dimg | first)) "docker_image_id" }}
{{- end -}}

{{- define "dimg_id" -}}
{{-   if eq (typeOf .) "chartutil.Values" -}}
{{-     $context := . -}}
{{      tuple $context | include "_dimg_id" }}
{{-   else if (ge (len .) 2) -}}
{{-     $name := index . 0 -}}
{{-     $context := index . 1 -}}
{{      tuple $name $context | include "_dimg_id2" }}
{{-   else -}}
{{-     $context := index . 0 -}}
{{      tuple $context | include "_dimg_id" }}
{{-   end -}}
{{- end -}}

{{- define "_werf_container_env" -}}
{{-   $context := index . 0 -}}
{{-   if $context.Values.global.werf.ci.is_branch -}}
- name: DOCKER_IMAGE_ID
  value: {{ tuple $context | include "_dimg_id" }}
{{-   end -}}
{{- end -}}

{{- define "_werf_container_env2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   if $context.Values.global.werf.ci.is_branch -}}
- name: DOCKER_IMAGE_ID
  value: {{ tuple $name $context | include "_dimg_id2" }}
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
