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

        def builder_checksum
          application.builder.before_install_checksum
        end

        def context
          [builder_checksum]
        end

        def prepare_image
          super
          application.builder.before_install(image)
        end

        alias dependencies context
      end # BeforeInstall
    end # Stage
  end # Build
end # Dapp
