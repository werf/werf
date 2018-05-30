module Dapp
  module Dimg
    module Dapp
      module Command
        module FlushLocal
          def flush_local
            lock("#{name}.images") do
              log_step_with_indent(:flush) do
                remove_project_images(dapp_project_dimgs, force: true)
                stages_flush_local if with_stages?
              end
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
