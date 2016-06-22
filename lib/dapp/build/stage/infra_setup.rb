module Dapp
  module Build
    module Stage
      class InfraSetup < Base
        def name
          :infra_setup
        end

        def image
          super do |image|
            build.infra_setup_do(image)
          end
        end

        def signature
          hashsum build.stages[:source_2].signature
        end
      end # InfraSetup
    end # Stage
  end # Build
end # Dapp
