module Dapp
  module Dimg
    module Dapp
      module Command
        module Cleanup
          def cleanup
            lock("#{name}.images") do
              log_step_with_indent(name) do
                dapp_containers_flush
                dapp_dangling_images_flush
                remove_images_by_query([
                  "#{host_docker_bin} images",
                  %(--format '{{if ne "#{stage_cache}" .Repository }}{{.Repository}}:{{.Tag}}{{ end }}'),
                  %(-f "label=dapp=#{stage_dapp_label}")
                ].join(' ')) # FIXME: negative filter is not currently supported by the Docker CLI
              end
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
