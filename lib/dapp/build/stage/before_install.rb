module Dapp
  module Build
    module Stage
      # BeforeInstall
      class BeforeInstall < Base
        def initialize(application, next_stage)
          @prev_stage = From.new(application, self)
          super
        end

        def empty?
          super && !application.builder.before_install?
        end

        def dependencies
          [application.builder.before_install_checksum]
        end

        def image
          super do |image|
            application.builder.before_install(image)
          end
        end
      end # BeforeInstall
    end # Stage
  end # Build
end # Dapp
