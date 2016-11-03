module Dapp
  # Project
  class Project
    # Command
    module Command
      # List
      module List
        def list
          build_configs.each do |config|
            if config._name.nil?
              log_warning("Project '#{name}' with nameless dimg!")
            else
              puts config._name
            end
          end
        end
      end
    end
  end # Project
end # Dapp
