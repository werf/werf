module Dapp
  module Kube
    module Dapp
      module Command
        module Deploy
          def kube_deploy
            helm_release do |release|
              do_deploy = proc do
                kube_flush_hooks_jobs(release)
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

              watch_hooks_thr.kill if watch_hooks_thr.alive?

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
