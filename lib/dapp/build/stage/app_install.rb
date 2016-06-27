module Dapp
  module Build
    module Stage
      class AppInstall < Base
        def initialize(build, relative_stage)
          @prev_stage = Source1.new(build, self)
          super
        end

        def image
          super do |image|
            build.app_install_do(image)
          end
        end
      end # AppInstall
    end # Stage
  end # Build
end # Dapp
