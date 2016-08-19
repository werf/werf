module Dapp
  module Build
    module Stage
      # GAPostInstallPatch
      class GAPostInstallPatch < GABase
        def initialize(application, next_stage)
          @prev_stage = GAPostInstallPatchDependencies.new(application, self)
          super
        end

        def prev_g_a_stage
          super.prev_stage
        end
      end # GAPostInstallPatch
    end # Stage
  end # Build
end # Dapp
