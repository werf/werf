module Dapp
  module Build
    module Stage
      module SetupGroup
        # Setup
        class Setup < Base
          include Mod::Group

          def initialize(application, next_stage)
            @prev_stage = GAPreSetupPatch.new(application, self)
            super
          end

          def empty?
            super && !application.builder.setup?
          end

          def dependencies
            prev_stage.prev_stage.dependencies # GAPreSetupPatchDependencies
          end

          def image
            super do |image|
              application.builder.setup(image)
            end
          end
        end # Setup
      end
    end # Stage
  end # Build
end # Dapp
