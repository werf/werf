module Dapp
  # Project
  class Project
    # Command
    module Command
      # Common
      module Common
        protected

        def project_images_names
          shellout!(%(docker images --format="{{.Repository}}:{{.Tag}}" #{stage_cache})).stdout.lines.map(&:strip)
        end

        def project_images_ids
          shellout!(%(docker images #{stage_cache} -q --no-trunc)).stdout.lines.map(&:strip)
        end

        def project_containers_flush
          remove_containers_by_query(%(docker ps -a -f "label=dapp" -f "name=#{container_name_prefix}" -q), force: true)
        end

        def project_dangling_images_flush
          remove_images_by_query(%(docker images -f "dangling=true" -f "label=dapp=#{stage_dapp_label}" -q), force: true)
        end

        def remove_images_by_query(images_query, force: false)
          with_subquery(images_query) { |ids| remove_images(ids, force: force) }
        end

        def remove_images(ids, force: false)
          remove_base('docker rmi%{force_option} %{ids}', ids, force: force)
        end

        def remove_containers_by_query(containers_query, force: false)
          with_subquery(containers_query) { |ids| remove_containers(ids, force: force) }
        end

        def remove_containers(ids, force: false)
          remove_base('docker rm%{force_option} %{ids}', ids, force: force)
        end

        def remove_base(query_format, ids, force: false)
          return if ids.empty?
          force_option = force ? ' -f' : ''
          ids.each_slice(50) { |chunk| run_command(query_format % ({ force_option: force_option, ids: chunk.join(' ') })) }
        end

        def with_subquery(query)
          return if (res = shellout!(query).stdout.strip.lines.map(&:strip)).empty?
          yield(res)
        end

        def image_labels(image_id)
          Image::Stage.image_config_option(image_id: image_id, option: 'labels')
        end

        def run_command(cmd)
          log(cmd) if log_verbose? || dry_run?
          shellout!(cmd) unless dry_run?
        end

        def container_name_prefix
          name
        end

        def validate_repo_name(repo)
          raise(Error::Project, code: :repo_name_incorrect, data: { name: repo }) unless Dapp::DockerRegistry.repo_name?(repo)
        end

        def proper_cache_version?
          !!cli_options[:proper_cache_version]
        end

        def log_proper_cache(&blk)
          log_step_with_indent(:'proper cache', &blk)
        end

        def one_dimg!
          raise Error::Project, code: :command_unexpected_dimgs_number, data: { dimgs_names: build_configs.map(&:_name).join(' ') } unless build_configs.one?
        end

        def push_format(dimg_name)
          if dimg_name.nil?
            spush_format
          else
            '%{repo}:%{dimg_name}-%{tag}'
          end
        end

        def spush_format
          '%{repo}:%{tag}'
        end
      end
    end
  end # Project
end # Dapp
