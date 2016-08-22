module Dapp
  # Project
  class Project
    # Command
    module Command
      # Build
      module Build
        def build
          build_configs.each do |config|
            log_step(config._name)
            with_log_indent do
              Application.new(config: config, project: self).build!
            end
          end
        end
      end
    end
  end # Project
end # Dapp
