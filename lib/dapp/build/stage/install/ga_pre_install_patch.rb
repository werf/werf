module Dapp
  module Build
    module Stage
      module InstallGroup
        # GAPostInstallPatch
        class GAPreInstallPatch < GABase
          include Mod::Group

          def initialize(dimg, next_stage)
            @prev_stage = GAPreInstallPatchDependencies.new(dimg, self)
            super
          end

          def prev_g_a_stage
            dependencies_stage.prev_stage # GAArchive
          end
        end # GAPostInstallPatch
      end
    end # Stage
  end # Build
end # Dapp
