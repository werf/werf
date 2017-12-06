module Dapp
  module Dimg
    module Dapp
      module Command
        module CleanupRepo
          def cleanup_repo
            lock_repo(repo = option_repo) do
              registry = registry(repo)

              repo_detailed_dimgs_images(registry).select do |image|
                case image[:labels]['dapp-tag-scheme']
                  when 'git_tag', 'git_branch', 'git_commit' then true
                  else false
                end && !deployed_docker_images.include?([image[:dimg], image[:tag]].join(':'))
              end.tap do |dimgs_images|
                cleanup_repo_by_nonexistent_git_tag(registry, dimgs_images)
                cleanup_repo_by_nonexistent_git_branch(registry, dimgs_images)
                cleanup_repo_by_nonexistent_git_commit(registry, dimgs_images)
              end

              begin
                registry.reset_cache
                repo_dimgs      = repo_dimgs_images(registry)
                repo_dimgstages = repo_dimgstages_images(registry)
                repo_dimgstages_cleanup(registry, repo_dimgs, repo_dimgstages)
              end if with_stages?
            end
          end

          def cleanup_repo_by_nonexistent_git_tag(registry, dimgs_images)
            cleanup_repo_by_nonexistent_git_base(dimgs_images, 'git_tag') do |dimg_image|
              delete_repo_image(registry, dimg_image) unless git_local_repo.tags.include?(dimg_image[:tag])
            end
          end

          def cleanup_repo_by_nonexistent_git_branch(registry, dimgs_images)
            cleanup_repo_by_nonexistent_git_base(dimgs_images, 'git_branch') do |dimg_image|
              delete_repo_image(registry, dimg_image) unless git_local_repo.remote_branches.include?(dimg_image[:tag])
            end
          end

          def cleanup_repo_by_nonexistent_git_commit(registry, dimgs_images)
            cleanup_repo_by_nonexistent_git_base(dimgs_images, 'git_commit') do |dimg_image|
              delete_repo_image(registry, dimg_image) unless git_local_repo.commit_exists?(dimg_image[:tag])
            end
          end

          def cleanup_repo_by_nonexistent_git_base(dimgs_images, dapp_tag_scheme)
            log_step_with_indent(:"nonexistent #{dapp_tag_scheme.split('_').join(' ')}") do
              dimgs_images
                .select { |dimg_image| dimg_image[:labels]['dapp-tag-scheme'] == dapp_tag_scheme }
                .each { |dimg_image| yield dimg_image }
            end
          end

          def repo_detailed_dimgs_images(registry)
            repo_dimgs_images(registry).each do |dimg|
              image_history = registry.image_history(dimg[:tag], dimg[:dimg])
              dimg[:parent] = image_history['container_config']['Image']
              dimg[:labels] = image_history['config']['Labels']
            end
          end

          def deployed_docker_images
            # open kube client, get all pods and select containers' images
            ::Dapp::Kube::Kubernetes::Client.tap do |kube|
              config_file = kube.kube_config_path
              unless File.exist?(config_file)
                return []
              end
            end

            client = ::Dapp::Kube::Kubernetes::Client.new

            namespaces = []
            # check connectivity for 2 seconds
            begin
              namespaces = client.namespace_list(excon_parameters: {:connect_timeout => 30})
            rescue Excon::Error::Timeout
              raise Kube::Error::Base, code: :connect_timeout
            end

            # get images from containers from pods from all namespaces.
            @kube_images ||= namespaces['items'].map do |item|
              item['metadata']['name']
            end.map do |ns|
              client.with_namespace(ns) do
                client.pod_list['items'].map do |pod|
                  pod['spec']['containers'].map{ |cont| cont['image'] }
                end
              end
            end.flatten.uniq.select do |image|
              image.start_with?(option_repo)
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
