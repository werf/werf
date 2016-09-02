module Dapp
  # Project
  class Project
    # Command
    module Command
      # Common
      module Common
        protected

        def project_images(basename)
          shellout!(%(docker images --format="{{.Repository}}:{{.Tag}}" #{stage_cache(basename)})).stdout.strip
        end

        def project_containers_flush(basename)
          remove_containers_by_query(%(docker ps -a -f "label=dapp" -f "name=#{container_name(basename)}" -q), force: true)
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

        def run_command(cmd)
          if dry_run?
            log(cmd)
          else
            shellout!(cmd)
          end
        end

        def stage_cache(basename)
          cache_format % { application_name: basename }
        end

        def stage_dapp_label(basename)
          stage_dapp_label_format % { application_name: basename }
        end

        def container_name(basename)
          basename
        end

        def proper_cache_version?
          !!cli_options[:proper_cache_version]
        end

        def log_proper_cache(&blk)
          log_step_with_indent(:'proper cache', &blk)
        end
      end
    end
  end # Project
end # Dapp
