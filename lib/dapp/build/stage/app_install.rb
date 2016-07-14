module Dapp
  module Build
    module Stage
      # AppInstall
      class AppInstall < Base
        def initialize(application, next_stage)
          @prev_stage = Source1.new(application, self)
          super
        end

        def image
          super do |image|
            application.builder.app_install(image)
          end
        end
      end # AppInstall
    end # Stage
  end # Build
end # Dapp
