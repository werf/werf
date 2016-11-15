module Dapp
  module Build
    module Stage
      module SetupGroup
        # GAPreSetupPatchDependencies
        class GAPreSetupPatchDependencies < GADependenciesBase
          include Mod::Group

          def initialize(dimg, next_stage)
            @prev_stage = BeforeSetupArtifact.new(dimg, self)
            super
          end

          def dependencies
            next_stage.next_stage.context # Setup
          end
        end # GAPreSetupPatchDependencies
      end
    end # Stage
  end # Build
end # Dapp
