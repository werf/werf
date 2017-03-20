module Dapp
  module Dimg
    module Build
      module Stage
        module Install
          class GAPostInstallPatchDependencies < GADependenciesBase
            include Mod::Group

            def initialize(dimg, next_stage)
              @prev_stage = Install.new(dimg, self)
              super
            end

            def dependencies
              dimg.stage_by_name(:before_setup).context
            end
          end # GAPostInstallPatchDependencies
        end
      end # Stage
    end # Build
  end # Dimg
end # Dapp
