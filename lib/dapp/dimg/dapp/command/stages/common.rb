module Dapp
  module Dimg
    module Dapp
      module Command
        module Stages
          module Common
            protected

            def repo_detailed_dimgs_images(registry)
              repo_dimgs_images(registry).each do |dimg|
                image_history = registry.image_history(dimg[:tag], dimg[:dimg])
                dimg[:parent] = image_history['container_config']['Image']
                dimg[:labels] = image_history['config']['Labels']
              end
            end

            def repo_dimgs_images(registry)
              [].tap do |dimgs_images|
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

            def repo_dimgstages_images(registry)
              tags_to_repo_images(registry, registry.dimgstages_tags)
            end

            def tags_to_repo_images(registry, tags, dimg_name = nil)
              tags.map { |tag| repo_image_format(registry, tag, dimg_name) }.compact
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
              log([repo_image[:dimg], repo_image[:tag]].compact.join(':')) if dry_run? || log_verbose?
              registry.image_delete(repo_image[:tag], repo_image[:dimg])           unless dry_run?
            end

            def select_dapp_artifacts_ids(labels)
              labels.select { |k, _v| k.start_with?('dapp-artifact') }.values
            end

            def dapp_git_repositories
              @dapp_git_repositories ||= begin
                {}.tap do |repositories|
                  dimgs = build_configs.map { |config| dimg(config: config, ignore_git_fetch: true, ignore_signature_auto_calculation: true) }
                  dimgs.each do |dimg|
                    [dimg, dimg.artifacts]
                      .flatten
                      .map(&:git_artifacts)
                      .flatten
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
