module Dapp
  module Build
    module Stage
      module InstallGroup
        # GAPostPatchDependencies
        class GAPostPatchDependencies < GADependenciesBase
          include Mod::Group

          def initialize(application, next_stage)
            @prev_stage = Artifact.new(application, self)
            super
          end

          def dependencies
            [application.builder.before_setup_checksum]
          end
        end # GAPostPatchDependencies
      end
    end # Stage
  end # Build
end # Dapp
