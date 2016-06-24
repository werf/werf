module Dapp
  module Build
    module Stage
      class InfraInstall < Base
        def initialize(build, relative_stage)
          @prev_stage = Prepare.new(build, self)
          super
        end

        def signature
          image.signature
        end

        protected

        def image
          super do |image|
            build.infra_install_do(image)
          end
        end
      end # InfraInstall
    end # Stage
  end # Build
end # Dapp
