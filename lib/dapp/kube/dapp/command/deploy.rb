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

            # TODO: Перенести код процесса выката в Helm::Manager

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
                chart_path: kube_tmp_chart_path,
                set: self.options[:helm_set_options],
                values: [*kube_values_paths, *kube_tmp_chart_secret_values_paths],
                deploy_timeout: self.options[:timeout] || 300
              )

              kube_flush_hooks_jobs(release)
              kube_run_deploy(release)
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
              decoded_data = kube_helm_decode_json(secret, yaml_load_file(secret_values_file))
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

          def kube_flush_hooks_jobs(release)
            release.hooks.values
              .reject { |job| ['0', 'false'].include? job.annotations["dapp/recreate"].to_s }
              .select { |job| kube_job_list.include? job.name }
              .each do |job|
                log_process("Delete hooks job `#{job.name}` for release #{release.name}", short: true) { kube_delete_job!(job.name) }
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

          def kube_run_deploy(release)
            log_process("Deploy release #{release.name}") do
              release_exists = shellout("helm status #{release.name}").status.success?

              watch_hooks_by_type = release.jobs.values
                .reduce({}) do |res, job|
                  if job.annotations['dapp/watch-logs'].to_s == 'true'
                    job.annotations['helm.sh/hook'].to_s.split(',').each do |hook_type|
                      res[hook_type] ||= []
                      res[hook_type] << job
                    end
                  end

                  res
                end
                .tap do |res|
                  res.values.each do |jobs|
                    jobs.sort_by! {|job| job.annotations['helm.sh/hook-weight'].to_i}
                  end
                end

              watch_hooks = if release_exists
                watch_hooks_by_type['pre-upgrade'].to_a + watch_hooks_by_type['post-upgrade'].to_a
              else
                watch_hooks_by_type['pre-install'].to_a + watch_hooks_by_type['post-install'].to_a
              end

              watch_hooks_thr = Thread.new do
                watch_hooks.each {|job| Kubernetes::Manager::Job.new(self, job.name).watch_till_done!}
              end

              deployment_managers = release.deployments.values
                .map {|deployment| Kubernetes::Manager::Deployment.new(self, deployment.name)}

              deployment_managers.each(&:before_deploy)

              release.deploy!

              deployment_managers.each(&:after_deploy)

              begin
                ::Timeout::timeout(self.options[:timeout] || 300) do
                  watch_hooks_thr.join
                  deployment_managers.each {|deployment_manager| deployment_manager.watch_till_ready!}
                end
              rescue ::Timeout::Error
                watch_hooks_thr.kill if watch_hooks_thr.alive?
                raise Error::Base, code: :deploy_timeout
              end
            end
          end

          def kube_check_helm_chart!
            raise Error::Command, code: :project_helm_chart_not_found, data: { path: kube_chart_path } unless kube_chart_path.exist?
          end

          def kube_tmp_chart_secret_path(*path)
            kube_tmp_chart_path('decoded-secret', *path).tap { |p| p.parent.mkpath }
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

          def kube_helm_manager
            @kube_helm_manager ||= Helm::Manager.new(self)
          end
        end
      end
    end
  end
end
