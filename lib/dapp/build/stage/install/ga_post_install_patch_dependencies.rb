module Dapp
  module Build
    module Stage
      module InstallGroup
        # GAPostInstallPatchDependencies
        class GAPostInstallPatchDependencies < GADependenciesBase
          include Mod::Group

          def initialize(application, next_stage)
            @prev_stage = Install.new(application, self)
            super
          end

          def dependencies
            [application.builder.before_setup_checksum]
          end
        end # GAPostInstallPatchDependencies
      end
    end # Stage
  end # Build
end # Dapp
