module Dapp
  module Build
    module Stage
      # Install
      class Install < Base
        def initialize(application, next_stage)
          @prev_stage = Source1.new(application, self)
          super
        end

        def empty?
          super && !application.builder.install?
        end

        def dependencies
          prev_stage.prev_stage.dependencies
        end

        def image
          super do |image|
            application.builder.install(image)
          end
        end
      end # Install
    end # Stage
  end # Build
end # Dapp
