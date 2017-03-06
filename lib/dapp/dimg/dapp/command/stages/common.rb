module Dapp
  module Dimg
    class Dapp
      module Command
        module Stages
          module Common
            protected

            def registry(repo)
              validate_repo_name(repo)
              ::Dapp::Dimg::DockerRegistry.new(repo)
            end

            def repo_dimgs_and_cache(registry)
              format = proc do |arr|
                arr.map do |tag|
                  if (id = registry.image_id(tag)).nil?
                    log_warning(desc: { code: 'tag_ignored', data: { tag: tag } })
                  else
                    [tag, id]
                  end
                end.compact.to_h
              end
              dimgs, stages = registry.tags.partition { |tag| !tag.start_with?('dimgstage') }
              [format.call(dimgs), format.call(stages)]
            end

            def registry_tags(registry)
              registry.tags
            rescue Exception::Registry => e
              raise unless e.net_status[:code] == :no_such_dimg
              log_warning(desc: { code: :dimg_not_found_in_registry })
              []
            end

            def delete_repo_image(registry, image_tag)
              if dry_run?
                log(image_tag)
              else
                registry.image_delete(image_tag)
              end
            end

            def select_dapp_artifacts_ids(hash)
              hash.select { |k, _v| k.start_with?('dapp-artifact') }.values
            end

            def dapp_git_repositories
              @dapp_git_repositories ||= begin
                {}.tap do |repositories|
                  dimgs = build_configs.map { |config| Dimg.new(config: config, dapp: self, ignore_git_fetch: true) }
                  dimgs.each do |dimg|
                    [dimg, dimg.artifacts].flatten
                                          .map(&:git_artifacts).flatten
                                          .map { |ga_artifact| repositories[ga_artifact.full_name] = ga_artifact.repo }
                  end
                end
              end
            end

            def proper_repo_cache?
              !!cli_options[:proper_repo_cache]
            end

            def proper_git_commit?
              !!cli_options[:proper_git_commit]
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
