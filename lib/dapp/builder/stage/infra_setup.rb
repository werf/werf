module Dapp
  module Builder
    module Stage
      class InfraSetup < Base
        def initialize(application, relative_stage)
          @prev_stage = Source2.new(application, self)
          super
        end

        def image
          super do |image|
            application.builder.infra_setup(image)
          end
        end
      end # InfraSetup
    end # Stage
  end # Builder
end # Dapp
