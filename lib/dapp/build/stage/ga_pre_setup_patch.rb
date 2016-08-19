module Dapp
  module Build
    module Stage
      # GAPreSetupPatch
      class GAPreSetupPatch < GABase
        def initialize(application, next_stage)
          @prev_stage = GAPreSetupPatchDependencies.new(application, self)
          super
        end

        def next_g_a_stage
          super.next_stage
        end
      end # GAPreSetupPatch
    end # Stage
  end # Build
end # Dapp
