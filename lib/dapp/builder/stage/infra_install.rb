module Dapp
  module Builder
    module Stage
      class InfraInstall < Base
        def initialize(application, relative_stage)
          @prev_stage = Prepare.new(application, self)
          super
        end

        def signature
          hashsum [prev_stage.signature, *application.builder.infra_install_checksum]
        end

        def image
          super do |image|
            application.builder.infra_install(image)
          end
        end
      end # InfraInstall
    end # Stage
  end # Builder
end # Dapp
