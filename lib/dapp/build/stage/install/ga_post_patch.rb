module Dapp
  module Build
    module Stage
      module InstallGroup
        # GAPostPatch
        class GAPostPatch < GABase
          include Mod::Group

          def initialize(application, next_stage)
            @prev_stage = GAPostPatchDependencies.new(application, self)
            super
          end

          def prev_g_a_stage
            super.prev_stage
          end
        end # GAPostPatch
      end
    end # Stage
  end # Build
end # Dapp
