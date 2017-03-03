module Dapp
  module Dimg
    module Build
      module Stage
        module Setup
          class GAPreSetupPatch < GABase
            include Mod::Group

            def initialize(dimg, next_stage)
              @prev_stage = GAPreSetupPatchDependencies.new(dimg, self)
              super
            end

            def prev_g_a_stage
              super.prev_stage.prev_stage # GAPostInstallPatch
            end
          end # GAPrePatch
        end
      end # Stage
    end # Build
  end # Dimg
end # Dapp
