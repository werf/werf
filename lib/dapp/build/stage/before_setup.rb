module Dapp
  module Build
    module Stage
      # BeforeSetup
      class BeforeSetup < Base
        def initialize(application, next_stage)
          @prev_stage = AfterInstallArtifact.new(application, self)
          super
        end

        def empty?
          !application.builder.before_setup?
        end

        def builder_checksum
          application.builder.before_setup_checksum
        end

        def context
          [builder_checksum]
        end

        def prepare_image
          super
          application.builder.before_setup(image)
        end
      end # BeforeSetup
    end # Stage
  end # Build
end # Dapp
