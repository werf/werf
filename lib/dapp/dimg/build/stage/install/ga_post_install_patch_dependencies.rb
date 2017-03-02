module Dapp
  module Dimg
    module Build
      module Stage
        module InstallGroup
          # GAPostInstallPatchDependencies
          class GAPostInstallPatchDependencies < GADependenciesBase
            include Mod::Group

            def initialize(dimg, next_stage)
              @prev_stage = Install.new(dimg, self)
              super
            end

            def dependencies
              next_stage.next_stage.next_stage.context # BeforeSetup
            end
          end # GAPostInstallPatchDependencies
        end
      end # Stage
    end # Build
  end # Dimg
end # Dapp
