module Dapp
  module Build
    module Stage
      # Setup
      class Setup < Base
        def initialize(application, next_stage)
          @prev_stage = Source3.new(application, self)
          super
        end

        def image
          super do |image|
            application.builder.setup(image)
          end
        end
      end # Setup
    end # Stage
  end # Build
end # Dapp
