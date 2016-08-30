module Dapp
  # Project
  class Project
    # Command
    module Command
      # Mrproper
      module Mrproper
        def mrproper
          log_step_with_indent('mrproper') do
            if proper_all?
              log_step_with_indent('containers') { remove_containers_by_query('docker ps -a -f "label=dapp" -q', force: true) }
              log_step_with_indent('images') { remove_images(dapp_images.lines.map(&:strip), force: true) }
            elsif proper_cache_version?
              log_proper_cache do
                all_images = dapp_images
                proper_cache_images = proper_cache_all_images
                remove_images(all_images.lines.select { |id| !proper_cache_images.lines.include?(id) }.map(&:strip))
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

        def dapp_images
          shellout!('docker images --format="{{.Repository}}:{{.Tag}}" -f "label=dapp"').stdout.strip
        end

        def proper_cache_all_images
          shellout!([
            'docker images',
            '--format="{{.Repository}}:{{.Tag}}"',
            %(-f "label=dapp-cache-version=#{Dapp::BUILD_CACHE_VERSION}")
          ].join(' ')).stdout.strip
        end
      end
    end
  end # Project
end # Dapp
