module Dapp
  module Build
    module Stage
      # GAPostSetupPatch
      class GAPostSetupPatch < GABase
        def initialize(application, next_stage)
          @prev_stage = GAPostSetupPatchDependencies.new(application, self)
          super
        end

        def prev_g_a_stage
          super.prev_stage
        end

        def next_g_a_stage
          next_stage
        end
      end # GAPostSetupPatch
    end # Stage
  end # Build
end # Dapp
