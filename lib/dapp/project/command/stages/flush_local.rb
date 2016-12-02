module Dapp
  # Project
  class Project
    # Command
    module Command
      module Stages
        # FlushLocal
        module FlushLocal
          def stages_flush_local
            lock("#{name}.images") do
              log_step_with_indent(name) do
                project_containers_flush
                project_dangling_images_flush
                remove_images(project_images_names)
              end
            end
          end
        end
      end
    end
  end # Project
end # Dapp
