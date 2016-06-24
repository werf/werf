module Dapp
  module Build
    module Stage
      class Source2 < SourceBase
        def initialize(build, relative_stage)
          @prev_stage = AppInstall.new(build, self)
          super
        end

        protected

        def dependencies_checksum
          hashsum [prev_stage.signature,
                   *build.infra_setup_checksum]
        end
      end # Source2
    end # Stage
  end # Build
end # Dapp
