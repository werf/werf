module Dapp
  module Build
    module Stage
      module SetupGroup
        # GAPrePatch
        class GAPreSetupPatch < GABase
          include Mod::Group

          def initialize(application, next_stage)
            @prev_stage = GAPreSetupPatchDependencies.new(application, self)
            super
          end

          def prev_g_a_stage
            super.prev_stage.prev_stage # GAPostInstallPatch
          end

          def next_g_a_stage
            super.next_stage # GAPostSetupPatch || GAArtifactPatch
          end
        end # GAPrePatch
      end
    end # Stage
  end # Build
end # Dapp
