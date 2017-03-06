module Dapp
  module Dimg
    class Dapp
      module Command
        module Stages
          module FlushLocal
            def stages_flush_local
              lock("#{name}.images") do
                log_step_with_indent(name) do
                  dapp_containers_flush
                  dapp_dangling_images_flush
                  remove_images(dapp_images_names)
                end
              end
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
