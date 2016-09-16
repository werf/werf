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

        def context
          [application.builder.before_install_checksum]
        end

        def prepare_image
          super
          application.builder.before_install(image)
        end

        alias_method :dependencies, :context
      end # BeforeInstall
    end # Stage
  end # Build
end # Dapp
