module Dapp
  module Kube
    module Dapp
      module Command
        module Common
          def lock_helm_release(&blk)
            lock("helm_release.#{kube_release_name}", &blk)
          end

          def helm_release
            kube_check_helm!
            kube_check_helm_template_plugin!
            kube_check_helm_chart!

            repo = option_repo
            tag = begin
              raise ::Dapp::Error::Command, code: :expected_only_one_tag, data: { tags: option_tags.join(', ') } if option_tags.count > 1
              option_tags.first
            end
            validate_repo_name!(repo)
            validate_tag_name!(tag)

            with_kube_tmp_chart_dir do
              kube_copy_chart
              kube_helm_decode_secrets
              kube_generate_helm_chart_tpl

              release = Helm::Release.new(
                  self,
                  name: kube_release_name,
                  repo: repo,
                  docker_tag: tag,
                  namespace: kube_namespace,
                  kube_context: custom_kube_context,
                  chart_path: kube_chart_path_for_helm,
                  set: self.options[:helm_set_options],
                  values: [*kube_values_paths, *kube_tmp_chart_secret_values_paths],
                  deploy_timeout: self.options[:timeout] || 300,
                  without_registry: self.options[:without_registry],
              )

              yield release
            end
          end

          def kube_check_helm_chart!
            raise ::Dapp::Error::Command, code: :helm_directory_not_exist, data: { path: kube_chart_path } unless kube_chart_path.exist?
          end

          def kube_check_helm_template_plugin!
            unless shellout!("helm plugin list | awk '{print $1}'").stdout.lines.map(&:strip).any? { |plugin| plugin == 'template' }
              raise ::Dapp::Error::Command, code: :helm_template_plugin_not_installed, data: { path: kube_chart_path }
            end
          end

          def kube_copy_chart
            FileUtils.cp_r("#{kube_chart_path}/.", kube_chart_path_for_helm)
          end

          def kube_helm_decode_secrets
            if secret.nil?
              log_warning(desc: {
                code: :dapp_secret_key_not_found,
                data: {not_found_in: secret_key_not_found_in.join(', ')}
              }) if !kube_secret_values_paths.empty? || kube_chart_secret_path.directory?
            end

            kube_helm_decode_secret_files
            kube_helm_decode_secret_values
          end

          def kube_helm_decode_secret_values
            kube_secret_values_paths.each_with_index do |secret_values_file, index|
              decoded_data = kube_helm_decode_json(secret, yaml_load_file(secret_values_file))
              kube_tmp_chart_secret_values_paths[index].write(decoded_data.to_yaml)
            end
          end

          def kube_helm_decode_secret_files
            return unless kube_chart_secret_path.directory?
            Dir.glob(kube_chart_secret_path.join('**/*')).each do |entry|
              next if File.directory?(entry)

              secret_relative_path = Pathname(entry).subpath_of(kube_chart_secret_path)

              if secret.nil?
                File.open(kube_tmp_chart_secret_path(secret_relative_path), 'wb:ASCII-8BIT', 0600) {|f| f.write ""}
              else
                kube_secret_file_validate!(entry)
                secret_data = secret.extract(IO.binread(entry).chomp("\n"))
                File.open(kube_tmp_chart_secret_path(secret_relative_path), 'wb:ASCII-8BIT', 0600) {|f| f.write secret_data}
              end
            end
          end

          def kube_generate_helm_chart_tpl
            cont = <<-EOF
{{- define "dapp_secret_file" -}}
{{-   $relative_file_path := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   $context.Files.Get (print "#{kube_tmp_chart_secret_path.subpath_of(kube_chart_path_for_helm)}/" $relative_file_path) -}}
{{- end -}}

{{- define "_dimg" -}}
{{-   $context := index . 0 -}}
{{-   if not $context.Values.global.dapp.is_nameless_dimg -}}
{{-     required "No dimg specified for template" nil -}}
{{-   end -}}
{{    $context.Values.global.dapp.dimg.docker_image }}
{{- end -}}

{{- define "_dimg2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   if $context.Values.global.dapp.is_nameless_dimg -}}
{{-     required (printf "No dimg should be specified for template, got `%s`" $name) nil -}}
{{-   end -}}
{{    index (required (printf "Unknown dimg `%s` specified for template" $name) (pluck $name $context.Values.global.dapp.dimg | first)) "docker_image" }}
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

{{- define "_dapp_container__imagePullPolicy" -}}
{{-   $context := index . 0 -}}
{{-   if $context.Values.global.dapp.ci.is_branch -}}
imagePullPolicy: Always
{{-   end -}}
{{- end -}}

{{- define "_dapp_container__image" -}}
{{-   $context := index . 0 -}}
image: {{ tuple $context | include "_dimg" }}
{{- end -}}

{{- define "_dapp_container__image2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
image: {{ tuple $name $context | include "_dimg2" }}
{{- end -}}

{{- define "dapp_container_image" -}}
{{-   if eq (typeOf .) "chartutil.Values" -}}
{{-     $context := . -}}
{{      tuple $context | include "_dapp_container__image" }}
{{      tuple $context | include "_dapp_container__imagePullPolicy" }}
{{-   else if (ge (len .) 2) -}}
{{-     $name := index . 0 -}}
{{-     $context := index . 1 -}}
{{      tuple $name $context | include "_dapp_container__image2" }}
{{      tuple $context | include "_dapp_container__imagePullPolicy" }}
{{-   else -}}
{{-     $context := index . 0 -}}
{{      tuple $context | include "_dapp_container__image" }}
{{      tuple $context | include "_dapp_container__imagePullPolicy" }}
{{-   end -}}
{{- end -}}

{{- define "_dimg_id" -}}
{{-   $context := index . 0 -}}
{{-   if not $context.Values.global.dapp.is_nameless_dimg -}}
{{-     required "No dimg specified for template" nil -}}
{{-   end -}}
{{    $context.Values.global.dapp.dimg.docker_image_id }}
{{- end -}}

{{- define "_dimg_id2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   if $context.Values.global.dapp.is_nameless_dimg -}}
{{-     required (printf "No dimg should be specified for template, got `%s`" $name) nil -}}
{{-   end -}}
{{    index (required (printf "Unknown dimg `%s` specified for template" $name) (pluck $name $context.Values.global.dapp.dimg | first)) "docker_image_id" }}
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

{{- define "_dapp_container_env" -}}
{{-   $context := index . 0 -}}
{{-   if $context.Values.global.dapp.ci.is_branch -}}
- name: DOCKER_IMAGE_ID
  value: {{ tuple $context | include "_dimg_id" }}
{{-   end -}}
{{- end -}}

{{- define "_dapp_container_env2" -}}
{{-   $name := index . 0 -}}
{{-   $context := index . 1 -}}
{{-   if $context.Values.global.dapp.ci.is_branch -}}
- name: DOCKER_IMAGE_ID
  value: {{ tuple $name $context | include "_dimg_id2" }}
{{-   end -}}
{{- end -}}

{{- define "dapp_container_env" -}}
{{-   if eq (typeOf .) "chartutil.Values" -}}
{{-     $context := . -}}
{{      tuple $context | include "_dapp_container_env" }}
{{-   else if (ge (len .) 2) -}}
{{-     $name := index . 0 -}}
{{-     $context := index . 1 -}}
{{      tuple $name $context | include "_dapp_container_env2" }}
{{-   else -}}
{{-     $context := index . 0 -}}
{{      tuple $context | include "_dapp_container_env" }}
{{-   end -}}
{{- end -}}
            EOF
            kube_chart_path_for_helm('templates/_dapp_helpers.tpl').write(cont)
          end

          def kube_tmp_chart_secret_path(*path)
            kube_chart_path_for_helm('decoded-secret', *path).tap { |p| p.parent.mkpath }
          end

          def kube_values_paths
            self.options[:helm_values_options].map { |p| Pathname(p).expand_path }.each do |f|
              raise ::Dapp::Error::Command, code: :values_file_not_found, data: { path: f } unless f.file?
            end
          end

          def kube_tmp_chart_secret_values_paths
            @kube_tmp_chart_secret_values_paths ||= kube_secret_values_paths.each_with_index.map { |f, i| kube_chart_path_for_helm( "decoded-secret-values-#{i}.yaml") }
          end

          def kube_secret_values_paths
            @kube_chart_secret_values_files ||= [].tap do |files|
              files << kube_chart_secret_values_path if kube_chart_secret_values_path.file?
              files.concat(options[:helm_secret_values_options].map { |p| Pathname(p).expand_path }.each do |f|
                kube_secret_file_validate!(f)
              end)
            end
          end

          def kube_chart_secret_values_path
            kube_chart_path('secret-values.yaml').expand_path
          end

          def kube_check_helm!
            raise ::Dapp::Error::Command, code: :helm_not_found if shellout('which helm').exitstatus == 1
          end

          def kube_release_name
            ENV["DAPP_HELM_RELEASE_NAME"] || "#{name}-#{kube_namespace}"
          end

          { encode: :generate, decode: :extract }.each do |type, secret_method|
            define_method "kube_helm_#{type}_json" do |secret, json|
              change_json_value = proc do |value|
                case value
                when Array then value.map { |v| change_json_value.call(v) }
                when Hash then send(:"kube_helm_#{type}_json", secret, value)
                when '', nil then ''
                else
                  secret.nil? ? '' : secret.public_send(secret_method, value)
                end
              end

              json.each { |k, v| json[k] = change_json_value.call(v) }
            end
          end

          def kube_secret_file_validate!(file_path)
            raise ::Dapp::Error::Command, code: :secret_file_not_found, data: { path: File.expand_path(file_path) } unless File.exist?(file_path)
            raise ::Dapp::Error::Command, code: :secret_file_empty, data: { path: File.expand_path(file_path) }     if File.read(file_path).strip.empty?
          end

          def secret_key_should_exist!
            raise ::Dapp::Error::Command, code: :secret_key_not_found, data: { not_found_in: secret_key_not_found_in.join(', ') } if secret.nil?
          end

          def kube_chart_secret_path(*path)
            kube_chart_path(kube_chart_secret_dir_name, *path).expand_path
          end

          def kube_chart_secret_dir_name
            'secret'
          end

          def kube_chart_path(*path)
            self.path('.helm', *path).expand_path
          end

          def with_kube_tmp_chart_dir
            yield if block_given?
          ensure
            FileUtils.rm_rf(@kube_tmp_helm_chart_dir) if @kube_tmp_helm_chart_dir
          end

          def kube_chart_path_for_helm(*path)
            chart_dir = ENV['DAPP_HELM_CHART_DIR'] || begin
              @kube_tmp_helm_chart_dir ||= Dir.mktmpdir('dapp-helm-chart-', tmp_base_dir)
            end
            make_path(chart_dir, *path).expand_path.tap { |p| p.parent.mkpath }
          end

          def secret
            @secret ||= begin
              unless (secret_key = ENV['DAPP_SECRET_KEY'])
                secret_key_not_found_in << '`DAPP_SECRET_KEY`'

                file_path = path('.dapp_secret_key')
                if file_path.file?
                  secret_key = path('.dapp_secret_key').read.chomp
                else
                  secret_key_not_found_in << "`#{file_path}`"
                end
              end

              Secret.new(secret_key) if secret_key
            end
          end

          def secret_key_not_found_in
            @secret_key_not_found_in ||= []
          end

          def namespace_option
            options[:namespace].nil? ? nil : consistent_uniq_slugify(options[:namespace])
          end

          def context_option
            options[:context]
          end

          def custom_kube_context
            ENV["KUBECONTEXT"] || context_option
          end

          def kube_context
            custom_kube_context || kubernetes_config.current_context_name
          end

          def kube_namespace
            namespace_option ||
              kubernetes_config.namespace(kube_context) ||
                "default"
          end

          def kubernetes_config
            @kubernetes_config ||= Kubernetes::Config.new_auto
          end

          def kubernetes
            @kubernetes ||= Kubernetes::Client.new(
              kubernetes_config,
              kube_context,
              kube_namespace,
              timeout: options[:kubernetes_timeout],
            )
          end
        end
      end
    end
  end
end
