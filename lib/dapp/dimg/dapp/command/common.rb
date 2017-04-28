module Dapp
  module Dimg
    module Dapp
      module Command
        module Common
          protected

          def dapp_images_names
            shellout!(%(docker images --format="{{.Repository}}:{{.Tag}}" #{stage_cache})).stdout.lines.map(&:strip)
          end

          def dapp_images_ids
            shellout!(%(docker images #{stage_cache} -q --no-trunc)).stdout.lines.map(&:strip)
          end

          def dapp_containers_flush
            remove_containers_by_query(%(docker ps -a -f "label=dapp" -f "name=#{container_name_prefix}" -q), force: true)
          end

          def dapp_dangling_images_flush
            remove_images_by_query(%(docker images -f "dangling=true" -f "label=dapp=#{stage_dapp_label}" -q), force: true)
          end

          def remove_images_by_query(images_query, force: false)
            with_subquery(images_query) { |ids| remove_images(ids, force: force) }
          end

          def remove_images(ids, force: false)
            remove_base('docker rmi%{force_option} %{ids}', ids.uniq, force: force)
          end

          def remove_containers_by_query(containers_query, force: false)
            with_subquery(containers_query) { |ids| remove_containers(ids, force: force) }
          end

          def remove_containers(ids, force: false)
            remove_base('docker rm%{force_option} %{ids}', ids.uniq, force: force)
          end

          def remove_base(query_format, ids, force: false)
            return if ids.empty?
            force_option = force ? ' -f' : ''
            ids.each_slice(50) { |chunk| run_command(format(query_format, force_option: force_option, ids: chunk.join(' '))) }
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

          def validate_repo_name!(repo)
            raise Error::Command, code: :repo_name_incorrect, data: { name: repo } unless ::Dapp::Dimg::DockerRegistry.repo_name?(repo)
          end

          def validate_image_name!(image)
            raise Error::Command, code: :image_name_incorrect, data: { name: image } unless ::Dapp::Dimg::Image::Docker.image_name?(image)
          end

          def validate_tag_name!(tag)
            raise Error::Command, code: :tag_name_incorrect, data: { name: tag } unless ::Dapp::Dimg::Image::Docker.tag?(tag)
          end

          def proper_cache_version?
            !!options[:proper_cache_version]
          end

          def log_proper_cache(&blk)
            log_step_with_indent(:'proper cache', &blk)
          end

          def one_dimg!
            return if build_configs.one?
            raise Error::Command, code: :command_unexpected_dimgs_number, data: { dimgs_names: build_configs.map(&:_name).join(' ') }
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

          def option_repo
            unless options[:repo].nil?
              return "localhost:5000/#{name}" if options[:repo] == ':minikube'
              options[:repo]
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
