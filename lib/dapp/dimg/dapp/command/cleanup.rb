module Dapp
  module Dimg
    module Dapp
      module Command
        module Cleanup
          def cleanup
            log_step_with_indent(:cleanup) do
              dapp_containers_flush
              dapp_dangling_images_flush
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
