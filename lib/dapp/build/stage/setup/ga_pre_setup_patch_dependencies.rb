module Dapp
  module Build
    module Stage
      module SetupGroup
        # GAPreSetupPatchDependencies
        class GAPreSetupPatchDependencies < GADependenciesBase
          include Mod::Group

          def initialize(application, next_stage)
            @prev_stage = BeforeSetupArtifact.new(application, self)
            super
          end

          def dependencies
            next_stage.next_stage.dependencies # Setup
          end
        end # GAPreSetupPatchDependencies
      end
    end # Stage
  end # Build
end # Dapp
