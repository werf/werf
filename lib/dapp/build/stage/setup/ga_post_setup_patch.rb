module Dapp
  module Build
    module Stage
      module SetupGroup
        # GAPostSetupPatch
        class GAPostSetupPatch < GABase
          include Mod::Group

          def initialize(application, next_stage)
            @prev_stage = GAPostSetupPatchDependencies.new(application, self)
            super
          end

          def prev_g_a_stage
            super.prev_stage # GAPreSetupPatch
          end

          def next_g_a_stage
            next_stage # GALatestPatch
          end
        end # GAPostSetupPatch
      end
    end # Stage
  end # Build
end # Dapp
