module Dapp
  module Build
    module Stage
      # BeforeSetup
      class BeforeSetup < Base
        def initialize(application, next_stage)
          @prev_stage = Artifact.new(application, self)
          super
        end

        def empty?
          super && !application.builder.before_setup?
        end

        def dependencies
          prev_stage.prev_stage.prev_stage.dependencies # GAPostInstallPatchDependencies
        end

        def prepare_image
          super
          application.builder.before_setup(image)
        end
      end # BeforeSetup
    end # Stage
  end # Build
end # Dapp
