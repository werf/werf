module Dapp
  # Dapp
  class Dapp
    # Command
    module Command
      # Mrproper
      module Mrproper
        def mrproper
          log_step_with_indent(:mrproper) do
            raise Error::Dapp, code: :mrproper_required_option unless proper_all? || proper_dev_mode_cache? || proper_cache_version?

            dapp_dangling_images_flush

            if proper_all?
              flush_by_label('dapp')
            elsif proper_dev_mode_cache?
              flush_by_label('dapp-dev-mode')
            elsif proper_cache_version?
              log_proper_cache do
                proper_cache_images = proper_cache_all_images
                remove_images(dapp_images_by_label('dapp').select { |id| !proper_cache_images.include?(id) }.map(&:strip))
              end
            end
          end
        end

        protected

        def flush_by_label(label)
          log_step_with_indent(:containers) { dapp_containers_flush_by_label(label) }
          log_step_with_indent(:images) { dapp_images_flush_by_label(label) }
        end

        def proper_all?
          !!cli_options[:proper_all]
        end

        def proper_dev_mode_cache?
          !!cli_options[:proper_dev_mode_cache]
        end

        def dapp_containers_flush_by_label(label)
          remove_containers_by_query(%(docker ps -a -f "label=#{label}" -q), force: true)
        end

        def dapp_dangling_images_flush_by_label(label)
          remove_images_by_query(%(docker images -f "dangling=true" -f "label=#{label}" -q), force: true)
        end

        def dapp_images_flush_by_label(label)
          dapp_dangling_images_flush_by_label(label)
          remove_images(dapp_images_by_label(label), force: true)
        end

        def dapp_images_by_label(label)
          @dapp_images ||= shellout!(%(docker images -f "dangling=false" --format="{{.Repository}}:{{.Tag}}" -f "label=#{label}")).stdout.lines.map(&:strip)
        end

        def proper_cache_all_images
          shellout!([
            'docker images',
            '--format="{{.Repository}}:{{.Tag}}"',
            %(-f "label=dapp-cache-version=#{::Dapp::BUILD_CACHE_VERSION}" -f "dangling=false")
          ].join(' ')).stdout.lines.map(&:strip)
        end
      end
    end
  end # Dapp
end # Dapp
