module Dapp
  module Dimg
    module Dapp
      module Command
        module CleanupRepo
          GIT_TAGS_LIMIT_POLICY     = 10
          EXPIRY_DATE_PERIOD_POLICY = 60 * 60 * 24 * 30
          GIT_COMMITS_LIMIT_POLICY  = 50

          def cleanup_repo
            lock_repo(repo = option_repo) do
              log_step_with_indent(repo) do
                log_step_with_indent('Searching for images being used in kubernetes clusters in one of the kube-config contexts') do
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
              cleanup_repo_by_nonexistent_git_base(registry, detailed_dimgs_images_by_scheme, scheme) do |detailed_dimg_image|
                case scheme
                  when 'git_tag'    then consistent_git_tags.include?(detailed_dimg_image[:tag])
                  when 'git_branch' then consistent_git_remote_branches.include?(detailed_dimg_image[:tag])
                  when 'git_commit' then git_own_repo.commit_exists?(detailed_dimg_image[:tag])
                  else
                    raise
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

          def cleanup_repo_by_nonexistent_git_base(registry, repo_dimgs_images_by_scheme, dapp_tag_scheme)
            nonexist_repo_images = repo_dimgs_images_by_scheme[dapp_tag_scheme]
              .select { |dimg_image| dimg_image[:labels]['dapp-tag-scheme'] == dapp_tag_scheme }
              .select { |dimg_image| !(yield dimg_image) }

            log_step_with_indent(:"#{dapp_tag_scheme.split('_').join(' ')} nonexistent") do
              nonexist_repo_images.each { |repo_image| delete_repo_image(registry, repo_image) }
            end unless nonexist_repo_images.empty?
          end

          def cleanup_repo_by_policies(registry, detailed_dimgs_images_by_scheme)
            cleanup_repo_tags_by_policies(registry, detailed_dimgs_images_by_scheme)
            cleanup_repo_commits_by_policies(registry, detailed_dimgs_images_by_scheme)
          end

          def cleanup_repo_tags_by_policies(registry, detailed_dimgs_images_by_scheme)
            detailed_dimgs_images = detailed_dimgs_images_by_scheme['git_tag'].select { |dimg| consistent_git_tags.include?(dimg[:tag]) }
            sorted_detailed_dimgs_images = detailed_dimgs_images.sort_by { |dimg| dimg[:created_at] }.reverse

            expired_dimgs_images, not_expired_dimgs_images = sorted_detailed_dimgs_images.partition do |dimg_image|
              dimg_image[:created_at] < git_tag_expiry_date_policy
            end

            log_step_with_indent(:"git tag date policy (before #{DateTime.strptime(git_tag_expiry_date_policy.to_s, '%s')})") do
              expired_dimgs_images.each { |dimg| delete_repo_image(registry, dimg) }
            end unless expired_dimgs_images.empty?

            not_expired_dimgs_images
              .each_with_object({}) { |dimg, images_by_dimg| (images_by_dimg[dimg[:dimg]] ||= []) << dimg }
              .each do |dimg_name, images|
                next if images[git_tags_limit_policy..-1].nil?
                log_step_with_indent(:"git tag limit policy (> #{git_tags_limit_policy}) (`#{dimg_name || 'nameless'}` dimg)") do
                  images[git_tags_limit_policy..-1].each { |dimg| delete_repo_image(registry, dimg) }
                end
              end
          end

          def cleanup_repo_commits_by_policies(registry, detailed_dimgs_images_by_scheme)
            detailed_dimgs_images_by_scheme['git_commit']
              .select { |dimg| git_own_repo.commit_exists?(dimg[:tag]) }
              .each_with_object({}) { |dimg, images_by_dimg| (images_by_dimg[dimg[:dimg]] ||= []) << dimg }
              .each do |dimg_name, images|
                next if images[git_commits_limit_policy..-1].nil?
                log_step_with_indent(:"git commit limit policy (> #{git_commits_limit_policy}) (`#{dimg_name || 'nameless'}` dimg)") do
                  images[git_commits_limit_policy..-1].each { |dimg| delete_repo_image(registry, dimg) }
                end
              end
          end

          def git_tag_expiry_date_policy
            @git_tag_expiry_date_policy = begin
              expiry_date_period_policy = policy_value('EXPIRY_DATE_PERIOD_POLICY', default: EXPIRY_DATE_PERIOD_POLICY)
              Time.now.to_i - expiry_date_period_policy
            end
          end

          def git_tags_limit_policy
            @git_tag_limit_policy ||= policy_value('GIT_TAGS_LIMIT_POLICY', default: GIT_TAGS_LIMIT_POLICY)
          end

          def git_commits_limit_policy
            @git_commits_limit_policy ||= policy_value('GIT_COMMITS_LIMIT_POLICY', default: GIT_COMMITS_LIMIT_POLICY)
          end

          def policy_value(env_key, default:)
            return default if (val = ENV[env_key]).nil?

            if val.to_i.to_s == val
              val.to_i
            else
              log_warning("WARNING: `#{env_key}` value `#{val}` is ignored (using default value `#{default}`)!")
              default
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
