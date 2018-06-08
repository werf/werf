module Dapp
  module Dimg
    module Dapp
      module Command
        module CleanupRepo
          DATE_POLICY = Time.now.to_i - 60 * 60 * 24 * 30
          GIT_TAGS_LIMIT_POLICY = 10

          def cleanup_repo
            lock_repo(repo = option_repo) do
              log_step_with_indent(repo) do
                log_step_with_indent("Searching for images being used in kubernetes clusters in one of the kube-config contexts") do
                  deployed_docker_images.each do |deployed_img|
                    log(deployed_img)
                  end
                end

                registry = dimg_registry(repo)

                if git_own_repo_exist?
                  cleanup_repo_by_nonexistent_git_primitive(registry, actual_detailed_dimgs_images_by_scheme(registry))
                  cleanup_repo_by_policies(registry, actual_detailed_dimgs_images_by_scheme(registry))
                end

                begin
                  repo_dimgs      = repo_dimgs_images(registry)
                  repo_dimgstages = repo_dimgstages_images(registry)
                  repo_dimgstages_cleanup(registry, repo_dimgs, repo_dimgstages)
                end if with_stages?
              end
            end
          end

          def actual_detailed_dimgs_images_by_scheme(registry)
            {}.tap do |detailed_dimgs_images_by_scheme|
              tagging_schemes.each { |scheme| detailed_dimgs_images_by_scheme[scheme] = [] }
              repo_detailed_dimgs_images(registry).each do |image|
                image_repository = [option_repo, image[:dimg]].compact.join('/')
                image_name = [image_repository, image[:tag]].join(':')

                should_be_ignored = deployed_docker_images.include?(image_name)

                if should_be_ignored
                  log "Keep in repo image that is being used in kubernetes: #{image_name}"
                  next
                end

                (detailed_dimgs_images_by_scheme[image[:labels]['dapp-tag-scheme']] ||= []) << image
              end
            end
          end

          def cleanup_repo_by_nonexistent_git_primitive(registry, detailed_dimgs_images_by_scheme)
            %w(git_tag git_branch git_commit).each do |scheme|
              cleanup_repo_by_nonexistent_git_base(detailed_dimgs_images_by_scheme, scheme) do |detailed_dimg_image|
                delete_repo_image(registry, detailed_dimg_image) unless begin
                  case scheme
                    when 'git_tag'    then consistent_git_tags.include?(detailed_dimg_image[:tag])
                    when 'git_branch' then consistent_git_remote_branches.include?(detailed_dimg_image[:tag])
                    when 'git_commit' then git_own_repo.commit_exists?(detailed_dimg_image[:tag])
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
            @consistent_git_remote_branches ||= git_own_repo.remote_branches.map(&method(:consistent_uniq_slugify))
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
                detailed_dimgs_images_by_scheme[scheme].select do |dimg|
                  !dry_run? && begin
                    if scheme == 'git_tag'
                      consistent_git_tags.include?(dimg[:tag])
                    elsif scheme == 'git_commit'
                      git_own_repo.commit_exists?(dimg[:tag])
                    end
                  end
                end
              end
            end.tap do |detailed_dimgs_images|
              sorted_detailed_dimgs_images = detailed_dimgs_images.sort_by { |dimg| dimg[:created_at] }.reverse
              expired_dimgs_images, not_expired_dimgs_images = sorted_detailed_dimgs_images.partition do |dimg_image|
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
            @git_consistent_tags ||= git_own_repo.tags.map { |t| [consistent_uniq_slugify(t), t] }.to_h
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
                client = ::Dapp::Kube::Kubernetes::Client.new(config, context_name, "default")
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
                  arr << cronjob_images(client)
                  arr << daemonset_images(client)
                  arr << deployment_images(client)
                  arr << job_images(client)
                  arr << replicaset_images(client)
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
