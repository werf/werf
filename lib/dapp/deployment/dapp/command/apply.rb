module Dapp
  module Deployment
    module Dapp
      module Command
        module Apply
          def deployment_apply
            repo = option_repo
            image_version = options[:image_version]

            validate_repo_name!(repo)
            validate_tag_name!(image_version)

            log_process("Applying deployment #{deployment.name}", verbose: true) do
              with_log_indent do
                deployment.kube.delete_unknown_resources!

                deployment.to_kube_bootstrap_pods(repo, image_version).each do |name, spec|
                  next if deployment.kube.pod_succeeded?(name)
                  deployment.kube.delete_pod!(name) if deployment.kube.pod_exist?(name)
                  log_process(:bootstrap, verbose: true) do
                    with_log_indent do
                      deployment.kube.run_job!(spec, name)
                    end
                  end
                end

                deployment.to_kube_before_apply_job_pods(repo, image_version).each do |name, spec|
                  log_process(:before_apply_job, verbose: true) do
                    deployment.kube.delete_pod!(name) if deployment.kube.pod_exist?(name)
                    with_log_indent do
                      deployment.kube.run_job!(spec, name)
                    end
                  end
                end

                deployment.apps.each do |app|
                  log_process("Applying app #{app.name}", verbose: true) do
                    with_log_indent do
                      (app.kube.existing_deployments_names - app.to_kube_deployments(repo, image_version).keys).each do |deployment_name|
                        app.kube.delete_deployment!(deployment_name)
                      end

                      (app.kube.existing_services_names - app.to_kube_services.keys).each do |service_name|
                        app.kube.delete_service!(service_name)
                      end

                      app.to_kube_bootstrap_pods(repo, image_version).each do |name, spec|
                        next if app.kube.pod_succeeded?(name)
                        app.kube.delete_pod!(name) if app.kube.pod_exist?(name)
                        log_process(:bootstrap, verbose: true) do
                          with_log_indent do
                            app.kube.run_job!(spec, name)
                          end
                        end
                      end

                      app.to_kube_before_apply_job_pods(repo, image_version).each do |name, spec|
                        log_process(:before_apply_job, verbose: true) do
                          app.kube.delete_pod!(name) if app.kube.pod_exist?(name)
                          with_log_indent do
                            app.kube.run_job!(spec, name)
                          end
                        end
                      end

                      app.to_kube_deployments(repo, image_version).each do |name, spec|
                        app.kube.apply_deployment!(name, spec)
                      end

                      app.to_kube_services.each do |name, spec|
                        app.kube.apply_service!(name, spec)
                      end
                    end
                  end
                end
              end
            end
          end
        end
      end
    end
  end
end
