module Dapp
  module Dimg
    module Dapp
      module Command
        module Stages
          module Common
            protected

            def repo_detailed_dimgs_images(registry)
              repo_dimgs_images(registry).select do |dimg|
                begin
                  image_config = registry.image_config(dimg[:tag], dimg[:dimg])
                  dimg[:created_at] = Time.parse(image_config["created"]).to_i
                  dimg[:parent] = image_config["container_config"]["Image"]
                  dimg[:labels] = image_config["config"]["Labels"]
                  dimg[:labels]['dapp'] == name
                rescue ::Dapp::Dimg::Error::Registry => e
                  raise unless e.net_status[:data][:message].include?("MANIFEST_UNKNOWN")
                  log_warning "WARNING: Ignore dimg `#{dimg[:dimg]}` tag `#{dimg[:tag]}`: got manifest-invalid-error from docker registry: #{e.net_status[:data][:message]}"
                  false
                end
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
              begin
                id = registry.image_id(tag, dimg_name)

                if id.nil?
                  log_warning(desc: { code: 'tag_ignored', data: { tag: tag } })
                  return nil
                end

                return { dimg: dimg_name, tag: tag, id: id }
              rescue ::Dapp::Dimg::Error::Registry => e
                raise unless e.net_status[:data][:message].include?("MANIFEST_UNKNOWN")
                log_warning "WARNING: Ignore dimg `#{dimg_name}` tag `#{tag}`: got not-found-error from docker registry on get-image-manifest request: #{e.net_status[:data][:message]}"
                return nil
              end
            end

            def delete_repo_image(registry, repo_image)
              log([repo_image[:dimg], repo_image[:tag]].compact.join(':')) if dry_run? || log_verbose?
              unless dry_run?
                begin
                  registry.image_delete(repo_image[:tag], repo_image[:dimg])
                rescue ::Dapp::Dimg::Error::Registry => e
                  raise unless e.net_status[:data][:message].include?("MANIFEST_UNKNOWN")
                  log_warning "WARNING: Ignore dimg `#{repo_image[:dimg]}` tag `#{repo_image[:tag]}`: got not-found-error from docker registry on image-delete request: #{e.net_status[:data][:message]}"
                end
              end
            end

            def select_dapp_artifacts_ids(labels)
              labels.select { |k, _v| k.start_with?('dapp-artifact') }.values
            end

            def dapp_git_repositories
              @dapp_git_repositories ||= begin
                {}.tap do |repositories|
                  dimgs = build_configs.map { |config| dimg(config: config, ignore_signature_auto_calculation: true) }
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
