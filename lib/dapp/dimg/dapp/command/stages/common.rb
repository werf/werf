module Dapp
  module Dimg
    module Dapp
      module Command
        module Stages
          module Common
            protected

            def registry(repo)
              validate_repo_name!(repo)
              ::Dapp::Dimg::DockerRegistry.new(repo)
            end

            def repo_dimgs_images(registry)
              @repo_dimgs_images ||= [].tap do |dimgs_images|
                with_registry_wrapper do
                  {}.tap do |dimgs_tags|
                    dimgs_tags[nil] = registry.nameless_dimg_tags
                    dimgs_names.each do |dimg_name|
                      dimgs_tags[dimg_name] = registry.dimg_tags(dimg_name)
                    end unless nameless_dimg?
                  end.each do |dimg_name, tags|
                    dimgs_images.concat(tags_to_repo_images(registry, tags, dimg_name))
                  end
                end
              end
            end

            def repo_dimgstages_images(registry)
              with_registry_wrapper do
                tags_to_repo_images(registry, registry.dimgstages_tags)
              end
            end

            def tags_to_repo_images(registry, tags, dimg_name = nil)
              tags.map { |tag| repo_image_format(registry, tag, dimg_name) }.compact
            end

            def with_registry_wrapper
              yield
            rescue Exception::Registry => e
              raise unless e.net_status[:code] == :no_such_dimg
              log_warning(desc: { code: :dimg_not_found_in_registry })
              []
            end

            def repo_image_format(registry, tag, dimg_name = nil)
              if (id = registry.image_id(tag, dimg_name)).nil?
                log_warning(desc: { code: 'tag_ignored', data: { tag: tag } })
                nil
              else
                { dimg: dimg_name, tag: tag, id: id }
              end
            end

            def delete_repo_image(registry, repo_image)
              if dry_run?
                log(repo_image[:tag])
              else
                registry.image_delete(repo_image[:tag], repo_image[:dimg])
              end
            end

            def select_dapp_artifacts_ids(labels)
              labels.select { |k, _v| k.start_with?('dapp-artifact') }.values
            end

            def dapp_git_repositories
              @dapp_git_repositories ||= begin
                {}.tap do |repositories|
                  dimgs = build_configs.map { |config| Dimg.new(config: config, dapp: self, ignore_git_fetch: true) }
                  dimgs.each do |dimg|
                    [dimg, dimg.artifacts].flatten
                                          .map(&:git_artifacts).flatten
                                          .map { |git_artifact| repositories[dimgstage_g_a_commit_label(git_artifact.paramshash)] = git_artifact.repo }
                  end
                end
              end
            end

            def proper_repo_cache?
              !!options[:proper_repo_cache]
            end

            def proper_git_commit?
              !!options[:proper_git_commit]
            end

            def stages_cleanup_option?
              proper_git_commit? || proper_cache_version? || proper_repo_cache?
            end

            def log_proper_git_commit(&blk)
              log_step_with_indent(:'proper git commit', &blk)
            end

            def lock_repo(repo, *args, &blk)
              lock("repo.#{hashsum repo}", *args, &blk)
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
