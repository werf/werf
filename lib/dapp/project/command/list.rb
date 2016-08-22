module Dapp
  # Project
  class Project
    # Command
    module Command
      # List
      module List
        def list
          build_configs.each { |config| puts config._name }
        end
      end
    end
  end # Project
end # Dapp
