module Dapp
  module Build
    module Stage
      # Source2Dependencies
      class Source2Dependencies < SourceDependenciesBase
        def initialize(application, next_stage)
          @prev_stage = Install.new(application, self)
          super
        end

        def signature
          hashsum [super, *application.builder.infra_setup_checksum]
        end
      end # Source2Dependencies
    end # Stage
  end # Build
end # Dapp
