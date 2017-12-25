module Dapp
  module Dimg
    module Dapp
      module Command
        module Common
          protected

          def dapp_project_dimgstages
            dapp_project_images.select { |image| image[:dimgstage] }
          end

          def dapp_project_dimgs
            dapp_project_images.select { |image| image[:dimg] }
          end

          def dapp_project_image_by_id(image_id)
            dapp_project_images.find { |image| image[:id] == image_id }
          end

          def dapp_project_image_labels(image)
            dapp_project_image_inspect(image)['Config']['Labels']
          end

          def dapp_project_image_inspect(image)
            image[:inspect] ||= begin
              cmd = shellout!("#{::Dapp::Dapp.host_docker} inspect --type=image #{image[:id]}")
              Array(JSON.parse(cmd.stdout.strip)).first || {}
            end
          end

          def dapp_project_images
            @dapp_project_images ||= [].tap do |images|
              images.concat prepare_docker_images(stage_cache, dimgstage: true)
              images.concat prepare_docker_images('-f label=dapp-dimg=true', dimg: true)
            end
          end

          def prepare_docker_images(extra_args, **extra_fields)
            [].tap do |images|
              shellout!(%(#{host_docker} images --format="{{.ID}};{{.Repository}}:{{.Tag}};{{.CreatedAt}}" -f "dangling=false" -f "label=dapp=#{name}" --no-trunc #{extra_args}))
                .stdout
                .lines
                .map(&:strip)
                .each do |l|
                id, name, created_at = l.split(';')
                images << { id: id, name: name, created_at: Time.parse(created_at), **extra_fields }
              end
            end
          end

          def remove_project_images(project_images)
            update_project_images_cache(project_images)
            remove_images(project_images_to_delete(project_images))
          end

          def update_project_images_cache(project_images)
            dapp_project_images.delete_if { |image| project_images.include?(image) }
          end

          def project_images_to_delete(project_images)
            project_images.map { |image| image[:dangling] ? image[:id] : image[:name] }
          end

          def dapp_containers_flush
            remove_containers_by_query(%(#{host_docker} ps -a -f "label=dapp" -q --no-trunc))
          end

          def dapp_dangling_images_flush
            remove_images_by_query(%(#{host_docker} images -f "dangling=true" -f "label=dapp" -q --no-trunc))
          end

          def remove_images_by_query(images_query)
            with_subquery(images_query) { |ids| remove_images(ids) }
          end

          def remove_images(images_ids_or_names)
            images_ids_or_names.uniq!
            check_user_containers!(images_ids_or_names)
            remove_base("#{host_docker} rmi%{force_option} %{ids}", images_ids_or_names, force: false)
          end

          def check_user_containers!(images_ids)
            return if images_ids.empty?
            run_command(%(#{host_docker} ps -a -q #{images_ids.uniq.map { |image_id| "--filter=ancestor=#{image_id}" }.join(' ')} --no-trunc)).tap do |res|
              raise ::Dapp::Error::Command, code: :user_containers_detected, data: { ids: res.stdout.strip } if res && !res.stdout.strip.empty? && !dry_run?
            end
          end

          def remove_containers_by_query(containers_query)
            with_subquery(containers_query) { |ids| remove_containers(ids) }
          end

          def remove_containers(ids)
            remove_base("#{host_docker} rm%{force_option} %{ids}", ids.uniq, force: true)
          end

          def remove_base(query_format, ids, force: false)
            return if ids.empty?
            force_option = force ? ' -f' : ''
            log(ids.join("\n")) if log_verbose? || dry_run?
            ids.each_slice(50) { |chunk| run_command(format(query_format, force_option: force_option, ids: chunk.join(' '))) }
          end

          def with_subquery(query)
            return if (res = shellout!(query).stdout.strip.lines.map(&:strip)).empty?
            yield(res)
          end

          def run_command(cmd)
            shellout!(cmd) unless dry_run?
          end

          def container_name_prefix
            name
          end

          def validate_repo_name!(repo)
            raise ::Dapp::Error::Command, code: :repo_name_incorrect, data: { name: repo } unless ::Dapp::Dimg::DockerRegistry.repo_name?(repo)
          end

          def validate_image_name!(image)
            raise ::Dapp::Error::Command, code: :image_name_incorrect, data: { name: image } unless ::Dapp::Dimg::Image::Docker.image_name?(image)
          end

          def validate_tag_name!(tag)
            raise ::Dapp::Error::Command, code: :tag_name_incorrect, data: { name: tag } unless ::Dapp::Dimg::Image::Docker.tag?(tag)
          end

          def proper_cache_version?
            !!options[:proper_cache_version]
          end

          def log_proper_cache(&blk)
            log_step_with_indent(:'proper cache', &blk)
          end

          def log_proper_repo_cache(&blk)
            log_step_with_indent(:'proper repo cache', &blk)
          end

          def one_dimg!
            return if build_configs.one?
            raise ::Dapp::Error::Command, code: :command_unexpected_dimgs_number, data: { dimgs_names: build_configs.map(&:_name).join(' ') }
          end

          def push_format(dimg_name)
            if dimg_name.nil?
              spush_format
            else
              '%{repo}/%{dimg_name}:%{tag}'
            end
          end

          def spush_format
            '%{repo}:%{tag}'
          end

          def with_stages?
            !!options[:with_stages]
          end
        end
      end
    end
  end # Dimg
end # Dapp
