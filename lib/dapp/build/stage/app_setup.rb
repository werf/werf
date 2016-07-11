module Dapp
  module Build
    module Stage
      class AppSetup < Base
        def initialize(application, next_stage)
          @prev_stage = Source3.new(application, self)
          super
        end

        def image
          super do |image|
            application.builder.app_setup(image)
          end
        end
      end # AppSetup
    end # Stage
  end # Build
end # Dapp
