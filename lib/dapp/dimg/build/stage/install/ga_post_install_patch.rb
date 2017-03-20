module Dapp
  module Dimg
    module Build
      module Stage
        module Install
          class GAPostInstallPatch < GABase
            include Mod::Group

            def initialize(dimg, next_stage)
              @prev_stage = GAPostInstallPatchDependencies.new(dimg, self)
              super
            end
          end # GAPostInstallPatch
        end
      end # Stage
    end # Build
  end # Dimg
end # Dapp
