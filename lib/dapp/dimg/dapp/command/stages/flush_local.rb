module Dapp
  module Dimg
    module Dapp
      module Command
        module Stages
          module FlushLocal
            def stages_flush_local
              lock("#{name}.images") do
                dapp_project_containers_flush
                dapp_project_dangling_images_flush
                remove_project_images(dapp_project_images_ids)
              end
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
