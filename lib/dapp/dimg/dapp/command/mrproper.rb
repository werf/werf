module Dapp
  module Dimg
    module Dapp
      module Command
        module Mrproper
          # rubocop:disable Metrics/PerceivedComplexity, Metrics/CyclomaticComplexity
          def mrproper
            log_step_with_indent(:mrproper) do
              raise ::Dapp::Error::Command, code: :mrproper_required_option if mrproper_command_without_any_option?

              if proper_all?
                proper_all
              elsif proper_dev_mode_cache?
                proper_dev_mode_cache
              elsif proper_cache_version?
                proper_cache_version
              end

              dapp_dangling_images_flush_by_label('dapp')
              dapp_tagless_images_flush
            end
          end
          # rubocop:enable Metrics/PerceivedComplexity, Metrics/CyclomaticComplexity

          protected

          def mrproper_command_without_any_option?
            !(proper_all? || proper_dev_mode_cache? || proper_cache_version?)
          end

          def proper_all?
            !!options[:proper_all]
          end

          def proper_dev_mode_cache?
            !!options[:proper_dev_mode_cache]
          end

          def proper_all
            flush_by_label('dapp')
          end

          def proper_dev_mode_cache
            flush_by_label('dapp-dev-mode')
          end

          def flush_by_label(label)
            dapp_containers_flush_by_label(label)
            dapp_images_flush_by_label(label)
          end

          def dapp_images_flush_by_label(label)
            log_step_with_indent('proper images') do
              remove_images(dapp_images_names_by_label(label), force: true)
            end
          end

          def proper_cache_version
            log_proper_cache do
              proper_cache_all_images_names.tap do |proper_cache_images|
                remove_images(dapp_images_names_by_label('dapp').select { |image_name| !proper_cache_images.include?(image_name) }, force: true)
              end
            end
          end

          def proper_cache_all_images_names
            ruby2go_image_images([{ label: "dapp" }, { label: "dapp-cache-version=#{::Dapp::BUILD_CACHE_VERSION}" }], ignore_tagless: true).map { |i| i["RepoTags"].empty? ? i["Id"] : i["RepoTags"] }.flatten
          end

          def dapp_images_names_by_label(label)
            ruby2go_image_images([{ label: "dapp" }, { label: label }], ignore_tagless: true).map { |i| i["RepoTags"].empty? ? i["Id"] : i["RepoTags"] }.flatten
          end
        end
      end
    end
  end # Dimg
end # Dapp
