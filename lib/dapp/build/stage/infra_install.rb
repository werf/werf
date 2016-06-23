module Dapp
  module Build
    module Stage
      class InfraInstall < Base
        def initialize(build, relative_stage)
          @prev_stage = Prepare.new(build, self)
          super
        end

        def name
          :infra_install
        end

        def image
          super do |image|
            build.infra_install_do(image)
          end
        end

        def signature
          image.signature
        end
      end # InfraInstall
    end # Stage
  end # Build
end # Dapp
