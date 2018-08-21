module Dapp
  module Dimg
    module Dapp
      module Command
        module Common
          def dimgstage_push_tag_format
            'dimgstage-%{signature}'
          end

          protected

          def dapp_project_dimgstages
            dapp_project_images.select { |image| image[:dimgstage] }
          end

          def dapp_project_dimgs
            dapp_project_images.select { |image| image[:dimg] }
          end

          def dapp_project_image_by_id(image_id)
            dapp_project_images.find { |image| image["Id"] == image_id }
          end

          def dapp_project_images
            @dapp_project_images ||= [].tap do |images|
              images.concat prepare_docker_images(extra_filters: [{ reference: stage_cache }], extra_image_fields: { dimgstage: true })
              images.concat prepare_docker_images(extra_filters: [{ label: "dapp-dimg=true" }], extra_image_fields: { dimg: true })
            end
          end

          def prepare_docker_images(extra_filters: [], extra_image_fields: {})
            filters = []
            filters << { dangling: "false" }
            filters << { label: "dapp=#{name}" }
            filters.concat(extra_filters)

            ruby2go_image_images(filters, ignore_tagless: true).map { |image| image.merge(**extra_image_fields) }
          end

          def ruby2go_image_images(filters, ignore_tagless:)
            Image::Stage.ruby2go_command(self, command: :images, options: { filters: filters }).tap do |images|
              if ignore_tagless
                break images.select do |image|
                  image["RepoTags"].reject! { |tag| tag.split(":").last == '<none>' }
                  !image["RepoTags"].empty?
                end
              end
            end
          end

          def ruby2go_image_containers(filters)
            Image::Stage.ruby2go_command(self, command: :containers, options: { filters: filters })
          end

          def ruby2go_image_rm(ids, force: false)
            Image::Stage.ruby2go_command(self, command: :rm, options: { ids: ids, force: force })
          end

          def ruby2go_image_rmi(ids, force: false)
            Image::Stage.ruby2go_command(self, command: :rmi, options: { ids: ids, force: force })
          end

          def remove_project_images(project_images, force: false)
            update_project_images_cache(project_images)
            remove_images(project_images_to_delete(project_images), force: force)
          end

          def update_project_images_cache(project_images)
            array_hash_delete_if_by_id(dapp_project_images, project_images)
          end

          def array_hash_delete_if_by_id(project_images, *images)
            project_images.delete_if { |image| images.flatten.any? { |i| image["Id"] == i["Id"] } }
          end

          def project_images_to_delete(project_images)
            project_images.map { |i| i["RepoTags"] }.flatten.uniq
          end

          def dapp_containers_flush_by_label(label)
            log_proper_containers do
              containers = ruby2go_image_containers([{ name: "dapp.build.", label: label }])
              remove_containers(containers.map { |c| c["Id"] })
            end
          end

          def dapp_dangling_images_flush_by_label(label)
            log_proper_flush_dangling_images do
              images = ruby2go_image_images([{ dangling: "true", label: label }], ignore_tagless: false)
              remove_images(images.map { |i| i["Id"] })
            end
          end

          def dapp_tagless_images_flush
            remove_images begin
              ruby2go_image_images([{ dangling: "false", label: "dapp" }], ignore_tagless: false)
                .map { |image| image["RepoTags"].select { |tag| tag.split(":").last == "<none>" } }
                .flatten
            end
          end

          def remove_images(images_ids_or_names, force: false)
            proc = proc { |refs| ruby2go_image_rmi(refs, force: force) }
            images_ids_or_names = ignore_used_images(images_ids_or_names) unless force
            remove_base(images_ids_or_names, proc: proc)
          end

          def ignore_used_images(images_ids_or_names)
            not_used_images = proc do |*image_id_or_name, log: true|
              images     = image_id_or_name.flatten
              filters    = images.map { |ref| { ancestor: ref } }

              containers = ruby2go_image_containers(filters).select { |c| images.include?(c["Image"]) || images.include?(c["ImageID"]) } # ancestor filter matches containers based on its image or a descendant of it
              if containers.empty?
                true
              else
                log_info("Skip `#{images.join('`, `')}` (used by containers: #{ containers.map { |c| c["Id"] }.join(' ')})") if log
                false
              end
            end

            if not_used_images.call(images_ids_or_names, log: false)
              images_ids_or_names
            else
              images_ids_or_names.select(&not_used_images)
            end
          end

          def remove_containers(ids)
            proc = proc { |refs| ruby2go_image_rm(refs, force: true) }
            remove_base(ids, proc: proc)
          end

          def remove_base(ids, proc:)
            return              if ids.empty?
            ids.uniq!
            log(ids.join("\n")) if dry_run?
            proc.call(ids)      unless dry_run?
          end

          def dimg_import_export_base(should_be_built: true)
            repo = option_repo
            validate_repo_name!(repo)
            build_configs.each do |config|
              log_dimg_name_with_indent(config) do
                yield dimg(config: config, should_be_built: should_be_built)
              end
            end
          end

          def validate_repo_name!(repo)
            raise ::Dapp::Error::Command, code: :repo_name_incorrect, data: { name: repo } unless ::Dapp::Dimg::DockerRegistry.repo_name?(repo)
          end

          def validate_image_name!(image)
            raise ::Dapp::Error::Command, code: :image_name_incorrect, data: { name: image } unless ::Dapp::Dimg::Image::Stage.image_name?(image)
          end

          def validate_tag_name!(tag)
            raise ::Dapp::Error::Command, code: :tag_name_incorrect, data: { name: tag } unless ::Dapp::Dimg::Image::Stage.tag?(tag)
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

          def log_proper_containers(&blk)
            log_step_with_indent(:'proper containers', &blk)
          end

          def log_proper_flush_dangling_images(&blk)
            log_step_with_indent(:'proper dangling', &blk)
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

          def dimgstage_push_format
            "%{repo}:#{dimgstage_push_tag_format}"
          end

          def with_stages?
            !!options[:with_stages]
          end
        end
      end
    end
  end # Dimg
end # Dapp
