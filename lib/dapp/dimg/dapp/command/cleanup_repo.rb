module Dapp
  module Dimg
    module Dapp
      module Command
        module CleanupRepo
          DATE_POLICY = Time.now.to_i - 60 * 60 * 24 * 30
          GIT_TAGS_LIMIT_POLICY = 10

          def cleanup_repo
            lock_repo(repo = option_repo) do
              registry = dimg_registry(repo)

              cleanup_repo_by_nonexistent_git_primitive(registry, actual_detailed_dimgs_images_by_scheme(registry))
              cleanup_repo_by_policies(registry, actual_detailed_dimgs_images_by_scheme(registry))

              begin
                repo_dimgs      = repo_dimgs_images(registry)
                repo_dimgstages = repo_dimgstages_images(registry)
                repo_dimgstages_cleanup(registry, repo_dimgs, repo_dimgstages)
              end if with_stages?
            end
          end

          def actual_detailed_dimgs_images_by_scheme(registry)
            {}.tap do |detailed_dimgs_images_by_scheme|
              tagging_schemes.each { |scheme| detailed_dimgs_images_by_scheme[scheme] = [] }
              repo_detailed_dimgs_images(registry).each do |image|
                next unless repo_dimg_image_should_be_ignored?(image)
                (detailed_dimgs_images_by_scheme[image[:labels]['dapp-tag-scheme']] ||= []) << image
              end
            end
          end

          def repo_dimg_image_should_be_ignored?(image)
            image_repository = [option_repo, image[:dimg]].compact.join('/')
            image_name = [image_repository, image[:tag]].join(':')
            !deployed_docker_images.include?(image_name)
          end

          def cleanup_repo_by_nonexistent_git_primitive(registry, detailed_dimgs_images_by_scheme)
            %w(git_tag git_branch git_commit).each do |scheme|
              cleanup_repo_by_nonexistent_git_base(detailed_dimgs_images_by_scheme, scheme) do |detailed_dimg_image|
                delete_repo_image(registry, detailed_dimg_image) unless begin
                  case scheme
                    when 'git_tag'    then consistent_git_tags.include?(detailed_dimg_image[:tag])
                    when 'git_branch' then consistent_git_remote_branches.include?(detailed_dimg_image[:tag])
                    when 'git_commit' then git_local_repo.commit_exists?(detailed_dimg_image[:tag])
                    else
                      raise
                  end
                end
              end unless detailed_dimgs_images_by_scheme[scheme].empty?
            end
          end

          def consistent_git_tags
            git_tag_by_consistent_tag_name.keys
          end

          def consistent_git_remote_branches
            @consistent_git_remote_branches ||= git_local_repo.remote_branches.map(&method(:consistent_uniq_slugify))
          end

          def cleanup_repo_by_nonexistent_git_base(repo_dimgs_images_by_scheme, dapp_tag_scheme)
            return if repo_dimgs_images_by_scheme[dapp_tag_scheme].empty?
            log_step_with_indent(:"nonexistent #{dapp_tag_scheme.split('_').join(' ')}") do
              repo_dimgs_images_by_scheme[dapp_tag_scheme]
                .select { |dimg_image| dimg_image[:labels]['dapp-tag-scheme'] == dapp_tag_scheme }
                .each { |dimg_image| yield dimg_image }
            end
          end

          def cleanup_repo_by_policies(registry, detailed_dimgs_images_by_scheme)
            %w(git_tag git_commit).each_with_object([]) do |scheme, dimgs_images|
              dimgs_images.concat begin
                detailed_dimgs_images_by_scheme[scheme].map do |dimg|
                  dimg[:created_at] = begin
                    if scheme == 'git_tag'
                      git_local_repo.tag_at(dimg[:tag])
                    elsif scheme == 'git_commit'
                      git_local_repo.commit_at(dimg[:tag])
                    end
                  end
                  dimg
                end
              end
            end
              .sort_by { |dimg| dimg[:created_at] }
              .tap do |sorted_dimgs_images|
                expired_dimgs_images, not_expired_dimgs_images = sorted_dimgs_images.partition do |dimg_image|
                  dimg_image[:created_at] < DATE_POLICY
                end

                log_step_with_indent(:"date policy (before #{DateTime.strptime(DATE_POLICY.to_s, '%s')})") do
                  expired_dimgs_images.each { |dimg| delete_repo_image(registry, dimg) }
                end

                {}.tap do |images_by_dimg|
                  not_expired_dimgs_images.each { |dimg| (images_by_dimg[dimg[:dimg]] ||= []) << dimg }
                  images_by_dimg.each do |dimg, images|
                    log_step_with_indent(:"limit policy (> #{GIT_TAGS_LIMIT_POLICY}) (`#{dimg}`)") do
                      images[GIT_TAGS_LIMIT_POLICY..-1].each { |dimg| delete_repo_image(registry, dimg) }
                    end unless images[GIT_TAGS_LIMIT_POLICY..-1].nil?
                  end
                end
              end
          end

          def git_tag_by_consistent_git_tag(consistent_git_tag)
            git_tag_by_consistent_tag_name[consistent_git_tag]
          end

          def git_tag_by_consistent_tag_name
            @git_consistent_tags ||= git_local_repo.tags.map { |t| [consistent_uniq_slugify(t), t] }.to_h
          end

          def repo_detailed_dimgs_images(registry)
            repo_dimgs_images(registry).each do |dimg|
              image_history = registry.image_history(dimg[:tag], dimg[:dimg])
              dimg[:parent] = image_history['container_config']['Image']
              dimg[:labels] = image_history['config']['Labels']
            end
          end

          def deployed_docker_images
            return [] if without_kube?

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

          def without_kube?
            !!options[:without_kube]
          end
        end
      end
    end
  end # Dimg
end # Dapp
