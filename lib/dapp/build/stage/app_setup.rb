module Dapp
  module Build
    module Stage
      class AppSetup < Base
        def name
          :app_setup
        end

        def image
          super do |image|
            build.app_setup_do(image)
          end
        end

        def signature
          hashsum build.stages[:source_3].signature
        end
      end # AppSetup
    end # Stage
  end # Build
end # Dapp
