module Dapp
  module Build
    module Stage
      class InfraSetup < Base
        def initialize(build, relative_stage)
          @prev_stage = Source2.new(build, self)
          super
        end

        def image
          super do |image|
            build.infra_setup_do(image)
          end
        end
      end # InfraSetup
    end # Stage
  end # Build
end # Dapp
