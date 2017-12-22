module Dapp
  module Dimg
    module Dapp
      module Command
        module FlushLocal
          def flush_local
            lock("#{name}.images") do
              log_step_with_indent(:flush) do
                stages_flush_local if with_stages?
                remove_project_images(dapp_project_dimgs_ids)
              end
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
