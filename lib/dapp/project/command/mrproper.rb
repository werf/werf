module Dapp
  # Project
  class Project
    # Command
    module Command
      # Mrproper
      module Mrproper
        def mrproper
          log_step_with_indent(:mrproper) do
            if proper_all?
              log_step_with_indent(:containers) { dapp_containers_flush }
              log_step_with_indent(:images) { dapp_images_flush }
            elsif proper_cache_version?
              log_proper_cache do
                proper_cache_images = proper_cache_all_images
                remove_images(dapp_images.lines.select { |id| !proper_cache_images.lines.include?(id) }.map(&:strip))
              end
            else
              raise Error::Project, code: :mrproper_required_option
            end
          end
        end

        protected

        def proper_all?
          !!cli_options[:proper_all]
        end

        def dapp_containers_flush
          remove_containers_by_query('docker ps -a -f "label=dapp" -q', force: true)
        end

        def dapp_dangling_images_flush
          remove_images_by_query('docker images -f "dangling=true" -f "label=dapp" -q', force: true)
        end

        def dapp_images_flush
          dapp_dangling_images_flush
          remove_images(dapp_images.lines.map(&:strip), force: true)
        end

        def dapp_images
          @dapp_images ||= shellout!('docker images -f "dangling=false" --format="{{.Repository}}:{{.Tag}}" -f "label=dapp"').stdout.strip
        end

        def proper_cache_all_images
          shellout!([
            'docker images',
            '--format="{{.Repository}}:{{.Tag}}"',
            %(-f "label=dapp-cache-version=#{Dapp::BUILD_CACHE_VERSION}" -f "dangling=false")
          ].join(' ')).stdout.strip
        end
      end
    end
  end # Project
end # Dapp
