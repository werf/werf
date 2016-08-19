module Dapp
  module Build
    module Stage
      module SetupGroup
        # GAPrePatch
        class GAPrePatch < GABase
          include Mod::Group

          def initialize(application, next_stage)
            @prev_stage = GAPrePatchDependencies.new(application, self)
            super
          end

          def prev_g_a_stage
            super.prev_stage
          end

          def next_g_a_stage
            super.next_stage
          end
        end # GAPrePatch
      end
    end # Stage
  end # Build
end # Dapp
