module Dapp
  module Dimg
    module Dapp
      module Command
        module CleanupRepo
          def cleanup_repo
            git_repo_option = git_own_repo_exist? ? JSON.dump(git_own_repo.get_ruby2go_state_hash) : nil
            ruby2go_cleanup_command(:gc, ruby2go_cleanup_gc_options_dump, local_repo: git_repo_option)
          end

          def ruby2go_cleanup_gc_options_dump
            ruby2go_cleanup_common_repo_options.merge(
              mode: {
                with_stages: with_stages?
              },
              deployed_docker_images: deployed_docker_images
            ).tap do |data|
              break JSON.dump(data)
            end
          end

          def deployed_docker_images
            @deployed_docker_images ||= search_deployed_docker_images
          end

          def search_deployed_docker_images
            return [] if without_kube?

            config = ::Dapp::Kube::Kubernetes::Config.new_auto_if_available
            return [] if config.nil?

            config.context_names.
              map {|context_name|
                client = ::Dapp::Kube::Kubernetes::Client.new(
                  config,
                  context_name,
                  "default",
                  timeout: options[:kubernetes_timeout],
                )
                search_deployed_docker_images_from_kube(client)
              }.flatten.sort.uniq
          end

          def search_deployed_docker_images_from_kube(client)
            namespaces = []
            # check connectivity for 2 seconds
            begin
              namespaces = client.namespace_list(excon_parameters: {:connect_timeout => 30})
            rescue Excon::Error::Timeout
              raise ::Dapp::Error::Default, code: :kube_connect_timeout
            end

            # get images from containers from pods from all namespaces.
            namespaces['items'].map do |item|
              item['metadata']['name']
            end.map do |ns|
              [].tap do |arr|
                client.with_namespace(ns) do
                  arr << pod_images(client)
                  arr << deployment_images(client)
                  arr << replicaset_images(client)
                  arr << statefulset_images(client)
                  arr << daemonset_images(client)
                  arr << job_images(client)
                  arr << cronjob_images(client)
                  arr << replicationcontroller_images(client)
                end
              end
            end.flatten.select do |img|
              img.start_with? option_repo
            end.sort.uniq
          end

          # pod items[] spec containers[] image
          def pod_images(client)
            client.pod_list['items'].map do |item|
              item['spec']['containers'].map{ |cont| cont['image'] }
            end
          end

          # cronjob items[] spec jobTemplate spec template spec containers[] image
          def cronjob_images(client)
            client.cronjob_list['items'].map do |item|
              item['spec']['jobTemplate']['spec']['template']['spec']['containers'].map{ |cont| cont['image'] }
            end
          end

          # daemonsets   items[] spec template spec containers[] image
          def daemonset_images(client)
            client.daemonset_list['items'].map do |item|
              item['spec']['template']['spec']['containers'].map{ |cont| cont['image'] }
            end
          end

          # deployment   items[] spec template spec containers[] image
          def deployment_images(client)
            client.deployment_list['items'].map do |item|
              item['spec']['template']['spec']['containers'].map{ |cont| cont['image'] }
            end
          end

          # job          items[] spec template spec containers[] image
          def job_images(client)
            client.job_list['items'].map do |item|
              item['spec']['template']['spec']['containers'].map{ |cont| cont['image'] }
            end
          end

          # replicasets  items[] spec template spec containers[] image
          def replicaset_images(client)
            client.replicaset_list['items'].map do |item|
              item['spec']['template']['spec']['containers'].map{ |cont| cont['image'] }
            end
          end

           # replicasets  items[] spec template spec containers[] image
           def statefulset_images(client)
            client.statefulset_list['items'].map do |item|
              item['spec']['template']['spec']['containers'].map{ |cont| cont['image'] }
            end
          end

          # replicationcontroller    items[] spec template spec containers[] image
          def replicationcontroller_images(client)
            client.replicationcontroller_list['items'].map do |item|
              item['spec']['template']['spec']['containers'].map{ |cont| cont['image'] }
            end
          end

          def without_kube?
            !!options[:without_kube]
          end
        end
      end
    end
  end # Dimg
end # Dapp
