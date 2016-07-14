module Dapp
  module Build
    module Stage
      # Source2
      class Source2 < SourceBase
        def initialize(application, next_stage)
          @prev_stage = AppInstall.new(application, self)
          super
        end

        protected

        def dependencies_checksum
          hashsum [prev_stage.signature,
                   *application.builder.infra_setup_checksum]
        end
      end # Source2
    end # Stage
  end # Build
end # Dapp
