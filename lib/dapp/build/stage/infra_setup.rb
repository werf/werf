module Dapp
  module Build
    module Stage
      # InfraSetup
      class InfraSetup < Base
        def initialize(application, next_stage)
          @prev_stage = Source2.new(application, self)
          super
        end

        def empty?
          super && !application.builder.infra_setup?
        end

        def dependencies
          prev_stage.prev_stage.dependencies
        end

        def image
          super do |image|
            application.builder.infra_setup(image)
          end
        end
      end # InfraSetup
    end # Stage
  end # Build
end # Dapp
