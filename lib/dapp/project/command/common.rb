module Dapp
  # Project
  class Project
    # Command
    module Command
      # Common
      module Common
        protected

        def containers_flush(basename)
          remove_containers(%(docker ps -a -f "label=dapp" -f "name=#{container_name(basename)}" -q), force: true)
        end

        def remove_images(images_query, force: false)
          force_option = force ? ' -f' : ''
          with_subquery(images_query) { |ids| run_command(%(docker rmi#{force_option} #{ids.join(' ')})) }
        end

        def remove_containers(containers_query, force: false)
          force_option = force ? ' -f' : ''
          with_subquery(containers_query) { |ids| run_command(%(docker rm#{force_option} #{ids.join(' ')})) }
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
      end
    end
  end # Project
end # Dapp
