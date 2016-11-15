module Dapp
  module Build
    module Stage
      module InstallGroup
        # GAPreInstallPatchDependencies
        class GAPreInstallPatchDependencies < GADependenciesBase
          include Mod::Group

          def initialize(dimg, next_stage)
            @prev_stage = GAArchive.new(dimg, self)
            super
          end

          def dependencies
            next_stage.next_stage.context # Install
          end

          def empty?
            super || dependencies_empty?
          end
        end # GAPreInstallPatchDependencies
      end
    end # Stage
  end # Build
end # Dapp
