module Dapp
  module Kube
    module Dapp
      module Command
        module Deploy
          def kube_deploy
            # TODO: move setup_ssh_agent to golang
            setup_ssh_agent

            command = "deploy"

            # TODO: move option_tags logic to golang
            tag = begin
              raise ::Dapp::Error::Command, code: :expected_only_one_tag, data: { tags: option_tags.join(', ') } if option_tags.count > 1
              option_tags.first
            end

            # TODO: move project name logic to golang
            project_name = name.to_s

            # TODO: move project dir logic to golang
            project_dir = path.to_s

            # TODO: move release name logic to golang
            release_name = kube_release_name

            # TODO: move repo logic to golang
            repo = option_repo

            dimgs = self.build_configs.map do |config|
              d = self.dimg(config: config, ignore_signature_auto_calculation: true)
              {"Name" => d.name, "ImageTag" => tag, "Repo" => repo}
            end.uniq do |dimg|
              dimg["Name"]
            end

            res = ruby2go_deploy(
              "command" => command,
              "projectName" => project_name,
              "projectDir" => project_dir,
              "tag" => tag,
              "releaseName" => release_name,
              "repo" => repo,
              "dimgs" => JSON.dump(dimgs),
              "rubyCliOptions" => JSON.dump(self.options),
              options: { host_docker_config_dir: self.class.host_docker_config_dir },
            )

            raise ::Dapp::Error::Command, code: :ruby2go_deploy_command_failed, data: { command: command, message: res["error"] } unless res["error"].nil?
          end

          def kube_deploy_old
            setup_ssh_agent

            helm_release do |release|
              do_deploy = proc do
                init_kube_deploy_start_time
                kube_run_deploy(release)
              end

              if dry_run?
                do_deploy.call
              else
                lock_helm_release &do_deploy
              end
            end
          end

          def init_kube_deploy_start_time
            @kube_deploy_start_time ||= Time.now.to_datetime.rfc3339
          end

          def kube_deploy_start_time
            @kube_deploy_start_time
          end

          def kube_flush_hooks_jobs(release)
            all_jobs_names = kube_all_jobs_names

            all_pods_specs = kubernetes.pod_list["items"]
              .map {|spec| Kubernetes::Client::Resource::Pod.new(spec)}

            release.hooks.values
              .reject { |job| ['0', 'false'].include? job.annotations["dapp/recreate"].to_s }
              .select { |job| all_jobs_names.include? job.name }
              .each do |job|
                log_process("Delete hook job `#{job.name}` (dapp/recreate)", short: true) do
                  kube_delete_job!(job.name, all_pods_specs) unless dry_run?
                end
              end
          end

          def kube_all_jobs_names
            kubernetes.job_list['items'].map { |i| i['metadata']['name'] }
          end

          def kube_delete_job!(job_name, all_pods_specs)
            job_spec = Kubernetes::Client::Resource::Pod.new kubernetes.job(job_name)

            job_pods_specs = all_pods_specs
              .select do |pod|
                Array(pod.metadata['ownerReferences']).any? do |owner_reference|
                  owner_reference['uid'] == job_spec.metadata['uid']
                end
              end

            job_pods_specs.each do |job_pod_spec|
              kubernetes.delete_pod!(job_pod_spec.name)
            end

            # FIXME: https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/
            # FIXME: orphanDependents deprecated, should be propagationPolicy=Orphan. But it does not work.
            # FIXME: Also, kubectl uses the same: https://github.com/kubernetes/kubernetes/blob/master/pkg/kubectl/delete.go#L388
            # FIXME: https://github.com/kubernetes/kubernetes/issues/46659
            kubernetes.delete_job!(job_name, orphanDependents: false)
            loop do
              break unless kubernetes.job?(job_name)
              sleep 1
            end
          end

          def kube_helm_auto_purge_trigger_file_path(release_name)
            File.join(self.class.home_dir, "helm", release_name, "auto_purge_failed_release_on_next_deploy")
          end

          def kube_create_helm_auto_purge_trigger_file(release_name)
            FileUtils.mkdir_p File.dirname(kube_helm_auto_purge_trigger_file_path(release_name))
            FileUtils.touch kube_helm_auto_purge_trigger_file_path(release_name)
          end

          def kube_delete_helm_auto_purge_trigger_file(release_name)
            if File.exists? kube_helm_auto_purge_trigger_file_path(release_name)
              FileUtils.rm_rf kube_helm_auto_purge_trigger_file_path(release_name)
            end
          end

          def kube_run_deploy(release)
            log_process("Deploy release #{release.name}") do
              helm_status_res = shellout([
                "helm",
                ("--kube-context #{custom_kube_context}" if custom_kube_context),
                "status",
                release.name,
              ].compact.join(" "))

              release_status = nil
              if helm_status_res.status.success?
                status_line = helm_status_res.stdout.lines.find {|l| l.start_with? "STATUS: "}
                release_status = status_line.partition(": ")[2].strip if status_line
              end

              release_exists = nil

              if not helm_status_res.status.success?
                # Helm release is not exists.
                release_exists = false

                # Create purge-trigger for the next run.
                kube_create_helm_auto_purge_trigger_file(release.name)
              elsif ["FAILED", "PENDING_INSTALL"].include? release_status
                release_exists = true

                if File.exists? kube_helm_auto_purge_trigger_file_path(release.name)
                  log_process("Purge helm release #{release.name}") do
                    shellout!([
                      "helm",
                      ("--kube-context #{custom_kube_context}" if custom_kube_context),
                      "delete",
                      "--purge #{release.name}",
                    ].compact.join(" "))
                  end

                  # Purge-trigger file remains to exist
                  release_exists = false
                end
              else
                if File.exists? kube_helm_auto_purge_trigger_file_path(release.name)
                  log_warning "[WARN] Will not purge helm release #{release.name}: expected FAILED or PENDING_INSTALL release status, got #{release_status}"
                end

                release_exists = true
                kube_delete_helm_auto_purge_trigger_file(release.name)
              end

              kube_flush_hooks_jobs(release)

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

              watch_hooks_thr = nil
              watch_hooks_condition_mutex = ::Mutex.new
              watch_hooks_condition = ::ConditionVariable.new
              deploy_has_began = false

              unless dry_run? and watch_hooks.any?
                watch_hooks_thr = Thread.new do
                  watch_hooks_condition_mutex.synchronize do
                    while not deploy_has_began do
                      watch_hooks_condition.wait(watch_hooks_condition_mutex)
                    end
                  end

                  begin
                    watch_hooks.each do |job|
                      begin
                        if ENV["KUBEDOG"] != "0"
                          timeout = self.options[:timeout] || 300
                          tmp_dir = Dir.mktmpdir('dapp-ruby2go-', tmp_base_dir)
                          res = nil

                          begin
                            res = ruby2go_deploy_watcher(
                              {
                                "action" => "watch job",
                                "resourceName" => job.name,
                                "namespace" => release.namespace,
                                "timeout" => timeout,
                                "logsFromTime" => kube_deploy_start_time,
                                "kubeContext" => custom_kube_context,
                              },
                              tmp_dir: tmp_dir,
                            )
                          rescue ::Dapp::Dapp::Ruby2Go::Error => err
                            # ignore interrupt
                            if err.net_status[:data][:status_code] == 17
                              res = {}
                            else
                              raise
                            end
                          ensure
                            FileUtils.rmtree(tmp_dir)
                          end

                          if res["error"]
                            $stderr.puts(::Dapp::Dapp.paint_string(res["error"], :warning))
                          end
                        else
                          Kubernetes::Manager::Job.new(self, job.name).watch_till_done!
                        end
                      rescue ::Exception => e
                        raise
                      end
                    end # watch_hooks.each
                  rescue Kubernetes::Error::Default => e
                    # Default-ошибка -- это ошибка для пользователя которую перехватывает и
                    # показывает bin/dapp, а затем dapp выходит с ошибкой.
                    # Нельзя убивать родительский поток по Default-ошибке
                    # из-за того, что в этот момент в нем вероятно работает helm,
                    # а процесс деплоя в helm прерывать не стоит.
                    # Поэтому перехватываем и просто отображаем произошедшую
                    # ошибку для информации пользователю без завершения работы dapp.
                    $stderr.puts(::Dapp::Dapp.paint_string(::Dapp::Helper::NetStatus.message(e), :warning))
                  end
                end # Thread
              end # unless

              deployment_managers = release.deployments.values
                .map {|deployment| Kubernetes::Manager::Deployment.new(self, deployment.name)}

              deployment_managers.each(&:before_deploy)

              log_process("#{release_exists ? "Upgrade" : "Install"} helm release #{release.name}") do
                watch_hooks_condition_mutex.synchronize do
                  deploy_has_began = true
                  # Фактически гарантируется лишь вывод сообщения log_process перед выводом из потока watch_thr
                  watch_hooks_condition.signal
                end

                cmd_res = if release_exists
                  release.upgrade_helm_release
                else
                  release.install_helm_release
                end

                if cmd_res.error?
                  if cmd_res.stderr.end_with? "has no deployed releases\n"
                    log_warning "[WARN] Helm release #{release.name} is in improper state: #{cmd_res.stderr}"
                    log_warning "[WARN] Helm release #{release.name} will be removed with `helm delete --purge` on the next run of `dapp kube deploy`"

                    kube_create_helm_auto_purge_trigger_file(release.name)
                  end

                  watch_hooks_thr.kill if !dry_run? && watch_hooks_thr && watch_hooks_thr.alive?

                  raise ::Dapp::Error::Command, code: :kube_helm_failed, data: {output: (cmd_res.stdout + cmd_res.stderr).strip}
                else
                  kube_delete_helm_auto_purge_trigger_file(release.name)
                  log_info((cmd_res.stdout + cmd_res.stderr).strip)

                  watch_hooks_thr.join if !dry_run? && watch_hooks_thr && watch_hooks_thr.alive?
                end
              end

              deployment_managers.each(&:after_deploy)

              unless dry_run?
                begin
                  timeout = self.options[:timeout] || 300
                  ::Timeout::timeout(timeout) do
                    deployment_managers.each {|deployment_manager|
                      if deployment_manager.should_watch?
                        if ENV["KUBEDOG"] != "0"
                          res = ruby2go_deploy_watcher(
                            "action" => "watch deployment",
                            "resourceName" => deployment_manager.name,
                            "namespace" => release.namespace,
                            "timeout" => timeout,
                            "logsFromTime" => kube_deploy_start_time,
                            "kubeContext" => custom_kube_context,
                          )

                          if res["error"]
                            raise ::Dapp::Kube::Kubernetes::Error::Default.new data: {message: res["error"]}
                          end
                        else
                          deployment_manager.watch_till_ready!
                        end
                      end
                    }
                  end
                rescue ::Timeout::Error
                  raise ::Dapp::Error::Command, code: :kube_deploy_timeout
                end
              end
            end
          end
        end
      end
    end
  end
end
