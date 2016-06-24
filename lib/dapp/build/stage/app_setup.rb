module Dapp
  module Build
    module Stage
      class AppSetup < Base
        def initialize(build, relative_stage)
          @prev_stage = Source3.new(build, self)
          super
        end

        protected

        def image
          super do |image|
            build.app_setup_do(image)
          end
        end
      end # AppSetup
    end # Stage
  end # Build
end # Dapp
