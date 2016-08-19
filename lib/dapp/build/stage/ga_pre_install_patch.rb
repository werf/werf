module Dapp
  module Build
    module Stage
      # GAPreInstallPatch
      class GAPreInstallPatch < GABase
        def initialize(application, next_stage)
          @prev_stage = GAPreInstallPatchDependencies.new(application, self)
          super
        end

        def prev_g_a_stage
          dependencies_stage.prev_stage
        end

        def next_g_a_stage
          super.next_stage
        end
      end # GAPreInstallPatch
    end # Stage
  end # Build
end # Dapp
