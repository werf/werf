module Dapp
  module Build
    module Stage
      # BeforeInstall
      class BeforeInstall < Base
        def initialize(dimg, next_stage)
          @prev_stage = From.new(dimg, self)
          super
        end

        def empty?
          super && !dimg.builder.before_install?
        end

        def context
          [builder_checksum]
        end

        def builder_checksum
          dimg.builder.before_install_checksum
        end

        def prepare_image
          super
          dimg.builder.before_install(image)
        end

        alias dependencies context
      end # BeforeInstall
    end # Stage
  end # Build
end # Dapp
