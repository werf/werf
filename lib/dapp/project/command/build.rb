module Dapp
  # Project
  class Project
    # Command
    module Command
      # Build
      module Build
        def build
          setup_ssh_agent

          build_configs.each do |config|
            log_step_with_indent(config._name) do
              Dimg.new(config: config, project: self).build!
            end
          end
        end
      end
    end
  end # Project
end # Dapp
