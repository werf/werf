module Dapp
  module Builder
    module Stage
      class Source2 < SourceBase
        def initialize(application, relative_stage)
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
  end # Builder
end # Dapp
