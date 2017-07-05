module Dapp
  module Kube
    module Dapp
      module Command
        module Deploy
          def kube_deploy
            raise Error::Command, code: :helm_directory_not_exist unless kube_chart_path.exist?
            kube_check_helm!
            kube_check_helm_chart!

            repo = option_repo
            image_version = options[:image_version]
            validate_repo_name!(repo)
            validate_tag_name!(image_version)

            begin
              kube_copy_chart
              kube_helm_decode_secrets
              kube_generate_helm_chart_tpl

              additional_values = [].tap do |options|
                options.concat((kube_values_paths + kube_tmp_chart_secret_values_paths).map { |p| "--values #{p}" })
              end

              set_options = [].tap do |options|
                options << "--set global.dapp.repo=#{repo}"
                options << "--set global.dapp.image_version=#{image_version}"
                options << "--set global.namespace=#{kube_namespace}"
                options.concat(self.options[:helm_set_options].map { |opt| "--set #{opt}" })
              end

              kube_flush_hooks_jobs(additional_values, set_options)
              kube_run_deploy(additional_values, set_options)
            ensure
              FileUtils.rm_rf(kube_tmp_chart_path)
            end
          end

          def kube_copy_chart
            FileUtils.cp_r("#{kube_chart_path}/.", kube_tmp_chart_path)
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
              decoded_data = kube_helm_decode_json(secret, YAML::load(secret_values_file.read))
              kube_tmp_chart_secret_values_paths[index].write(decoded_data.to_yaml)
            end
          end

          def kube_helm_decode_secret_files
            return unless kube_chart_secret_path.directory?
            Dir.glob(kube_chart_secret_path.join('**/*')).each do |entry|
              next unless File.file?(entry)
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
{{- printf "%s:%s-%s" $context.Values.global.dapp.repo $name $context.Values.global.dapp.image_version -}}
{{- else -}}
{{- $context := index . 0 -}}
{{- printf "%s:%s" $context.Values.global.dapp.repo $context.Values.global.dapp.image_version -}}
{{- end -}}                                                                               
{{- end -}}                                                                                         

{{- define "dapp_secret_file" -}}
{{- $relative_file_path := index . 0 -}}
{{- $context := index . 1 -}}
{{- $context.Files.Get (print "#{kube_tmp_chart_secret_path.subpath_of(kube_tmp_chart_path)}/" $relative_file_path) -}}
{{- end -}}
            EOF
            kube_tmp_chart_path('templates/_dapp_helpers.tpl').write(cont)
          end

          def kube_flush_hooks_jobs(additional_values, set_options)
            return if (config_jobs_names = kube_helm_hooks_jobs_to_delete(additional_values, set_options).keys).empty?
            config_jobs_names.select { |name| kube_job_list.include? name }.each do |name|
              log_process("Delete hooks job `#{name}` for release #{kube_release_name} ", short: true) { kube_delete_job!(name) }
            end
          end

          def kube_helm_hooks_jobs_to_delete(additional_values, set_options)
            generator = proc do |text|
              text.split(/# Source.*|---/).reject {|c| c.strip.empty? }.map {|c| YAML::load(c) }.reduce({}) do |objects, c|
                objects[c['kind']] ||= {}
                objects[c['kind']][(c['metadata'] || {})['name']] = c
                objects
              end
            end

            args = [kube_release_name, kube_tmp_chart_path, additional_values, set_options, kube_helm_extra_options(dry_run: true)].flatten
            output = shellout!("helm upgrade #{args.join(' ')}").stdout

            manifest_start_index = output.lines.index("MANIFEST:\n") + 1
            hook_start_index     = output.lines.index("HOOKS:\n") + 1
            configs = generator.call(output.lines[hook_start_index..manifest_start_index-1].join)

            (configs['Job'] || {}).reject do |_, c|
              c['metadata'] ||= {}
              c['metadata']['annotations'] ||= {}
              c['metadata']['annotations']['helm.sh/resource-policy'] == 'keep'
            end
          end

          def kube_job_list
            kubernetes.job_list['items'].map { |i| i['metadata']['name'] }
          end

          def kube_delete_job!(name)
            kubernetes.delete_job!(name)
            loop do
              break unless kubernetes.job?(name)
              sleep 1
            end
          end

          def kube_run_deploy(additional_values, set_options)
            log_process("Deploy release #{kube_release_name}") do
              args = [kube_release_name, kube_tmp_chart_path, additional_values, set_options, kube_helm_extra_options].flatten
              kubernetes.create_namespace!(kube_namespace) unless kubernetes.namespace?(kube_namespace)
              shellout! "helm upgrade #{args.join(' ')}", verbose: true
            end
          end

          def kube_check_helm_chart!
            raise Error::Command, code: :project_helm_chart_not_found, data: { path: kube_chart_path } unless kube_chart_path.exist?
          end

          def kube_helm_extra_options(dry_run: dry_run?)
            [].tap do |options|
              options << "--namespace #{kube_namespace}"
              options << '--install'
              options << '--wait'
              options << '--timeout 1800'

              options << '--dry-run' if dry_run
              options << '--debug'   if dry_run || log_verbose?
            end
          end

          def kube_tmp_chart_secret_path(*path)
            kube_tmp_chart_path('decoded-secret', *path).tap { |p| p.parent.mkpath }
          end

          def kube_tmp_chart_path(*path)
            @kube_tmp_path ||= Dir.mktmpdir('dapp-helm-chart-', tmp_base_dir)
            make_path(@kube_tmp_path, *path).expand_path.tap { |p| p.parent.mkpath }
          end

          def kube_values_paths
            self.options[:helm_values_options].map { |p| Pathname(p).expand_path }.each do |f|
              raise Error::Command, code: :values_file_not_found, data: { path: f } unless f.file?
            end
          end

          def kube_tmp_chart_secret_values_paths
            @kube_tmp_chart_secret_values_paths ||= kube_secret_values_paths.map { |f| kube_tmp_chart_path("#{SecureRandom.uuid}-#{f.basename}") }
          end

          def kube_secret_values_paths
            @kube_chart_secret_values_files ||= [].tap do |files|
              files << kube_chart_secret_values_path if kube_chart_secret_values_path.file?
              files.concat(options[:helm_secret_values_options].map { |p| Pathname(p).expand_path }.each do |f|
                raise Error::Command, code: :secret_values_file_not_found, data: { path: f } unless f.file?
              end)
            end
          end

          def kube_chart_secret_values_path
            kube_chart_path('secret-values.yaml').expand_path
          end

          def kube_chart_secret_path(*path)
            kube_chart_path('secret', *path).expand_path
          end
        end
      end
    end
  end
end
