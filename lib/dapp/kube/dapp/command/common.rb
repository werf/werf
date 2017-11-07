module Dapp
  module Kube
    module Dapp
      module Command
        module Common
          def helm_release
            kube_check_helm!
            kube_check_helm_template_plugin!
            kube_check_helm_chart!

            repo = option_repo
            image_version = options[:image_version]
            validate_repo_name!(repo)
            validate_tag_name!(image_version)

            with_kube_tmp_chart_dir do
              kube_copy_chart
              kube_helm_decode_secrets
              kube_generate_helm_chart_tpl

              release = Helm::Release.new(
                  self,
                  name: kube_release_name,
                  repo: repo,
                  image_version: image_version,
                  namespace: kube_namespace,
                  chart_path: kube_chart_path_for_helm,
                  set: self.options[:helm_set_options],
                  values: [*kube_values_paths, *kube_tmp_chart_secret_values_paths],
                  deploy_timeout: self.options[:timeout] || 300
              )

              yield release
            end
          end

          def kube_check_helm_chart!
            raise Error::Command, code: :helm_directory_not_exist, data: { path: kube_chart_path } unless kube_chart_path.exist?
          end

          def kube_check_helm_template_plugin!
            unless shellout!("helm plugin list | awk '{print $1}'").stdout.lines.map(&:strip).any? { |plugin| plugin == 'template' }
              raise Error::Command, code: :helm_template_plugin_not_installed, data: { path: kube_chart_path }
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
            else
              kube_helm_decode_secret_files
            end
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
              kube_secret_file_validate!(entry)
              secret_relative_path = Pathname(entry).subpath_of(kube_chart_secret_path)
              secret_data = secret.extract(IO.binread(entry).chomp("\n"))
              File.open(kube_tmp_chart_secret_path(secret_relative_path), 'wb:ASCII-8BIT', 0400) {|f| f.write secret_data}
            end
          end

          def kube_generate_helm_chart_tpl
            cont = <<-EOF
{{/* vim: set filetype=mustache: */}}

{{- define "dimg" -}}
{{- if (ge (len (index .)) 2) -}}
{{- $name := index . 0 -}}
{{- $context := index . 1 -}}
{{- printf "%v:%v-%v" $context.Values.global.dapp.repo $name $context.Values.global.dapp.image_version -}}
{{- else -}}
{{- $context := index . 0 -}}
{{- printf "%v:%v" $context.Values.global.dapp.repo $context.Values.global.dapp.image_version -}}
{{- end -}}
{{- end -}}

{{- define "dapp_secret_file" -}}
{{- $relative_file_path := index . 0 -}}
{{- $context := index . 1 -}}
{{- $context.Files.Get (print "#{kube_tmp_chart_secret_path.subpath_of(kube_chart_path_for_helm)}/" $relative_file_path) -}}
{{- end -}}
            EOF
            kube_chart_path_for_helm('templates/_dapp_helpers.tpl').write(cont)
          end

          def kube_tmp_chart_secret_path(*path)
            kube_chart_path_for_helm('decoded-secret', *path).tap { |p| p.parent.mkpath }
          end

          def kube_values_paths
            self.options[:helm_values_options].map { |p| Pathname(p).expand_path }.each do |f|
              raise Error::Command, code: :values_file_not_found, data: { path: f } unless f.file?
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
            raise Error::Command, code: :helm_not_found if shellout('which helm').exitstatus == 1
          end

          def kube_release_name
            "#{name}-#{kube_namespace}".slugify
          end

          def kube_namespace
            kubernetes.namespace
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
            raise Error::Command, code: :secret_file_not_found, data: { path: File.expand_path(file_path) } unless File.exist?(file_path)
            raise Error::Command, code: :secret_file_empty, data: { path: File.expand_path(file_path) } if File.read(file_path).strip.empty?
          end

          def secret_key_should_exist!
            raise(Error::Command,
              code: :secret_key_not_found,
              data: {not_found_in: secret_key_not_found_in.join(', ')}
            ) if secret.nil?
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

                if dappfile_exists?
                  file_path = path('.dapp_secret_key')
                  if file_path.file?
                    secret_key = path('.dapp_secret_key').read.chomp
                  else
                    secret_key_not_found_in << "`#{file_path}`"
                  end
                else
                  log_warning(desc: { code: :secret_key_dappfile_not_found })
                end
              end

              Secret.new(secret_key) if secret_key
            end
          end

          def secret_key_not_found_in
            @secret_key_not_found_in ||= []
          end

          def kubernetes
            @kubernetes ||= begin
              namespace = options[:namespace].nil? ? nil : options[:namespace].tr('_', '-')
              Kubernetes::Client.new(namespace: namespace)
            end
          end
        end
      end
    end
  end
end
