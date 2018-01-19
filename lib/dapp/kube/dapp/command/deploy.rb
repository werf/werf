module Dapp
  module Kube
    module Dapp
      module Command
        module Deploy
          def kube_deploy
            helm_release do |release|
              do_deploy = proc do
                kube_run_deploy(release)
              end

              if dry_run?
                do_deploy.call
              else
                lock_helm_release &do_deploy
              end
            end
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

          def kube_run_deploy(release)
            log_process("Deploy release #{release.name}") do
              kube_flush_hooks_jobs(release)

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
                      Kubernetes::Manager::Job.new(self, job.name).watch_till_done!
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

              log_process("Run helm") do
                watch_hooks_condition_mutex.synchronize do
                  deploy_has_began = true
                  # Фактически гарантируется лишь вывод сообщения log_process перед выводом из потока watch_thr
                  watch_hooks_condition.signal
                end

                cmd_res = release.helm_upgrade!

                if cmd_res.error?
                  raise ::Dapp::Error::Command, code: :kube_helm_failed, data: {output: (cmd_res.stdout + cmd_res.stderr).strip}
                else
                  watch_hooks_thr.join if !dry_run? && watch_hooks_thr && watch_hooks_thr.alive?
                  log_info((cmd_res.stdout + cmd_res.stderr).strip)
                end
              end

              deployment_managers.each(&:after_deploy)

              unless dry_run?
                begin
                  ::Timeout::timeout(self.options[:timeout] || 300) do
                    deployment_managers.each {|deployment_manager| deployment_manager.watch_till_ready!}
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
