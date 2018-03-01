module Dapp
  module Dimg
    module Dapp
      module Command
        module Cleanup
          def cleanup
            log_step_with_indent(:cleanup) do
              dapp_containers_flush_by_label('dapp')
              dapp_dangling_images_flush_by_label('dapp')
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
