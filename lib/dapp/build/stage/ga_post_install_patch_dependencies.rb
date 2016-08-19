module Dapp
  module Build
    module Stage
      # GAPostInstallPatchDependencies
      class GAPostInstallPatchDependencies < GADependenciesBase
        def initialize(application, next_stage)
          @prev_stage = Artifact.new(application, self)
          super
        end

        def dependencies
          [application.builder.before_setup_checksum]
        end
      end # GAPostInstallPatchDependencies
    end # Stage
  end # Build
end # Dapp
