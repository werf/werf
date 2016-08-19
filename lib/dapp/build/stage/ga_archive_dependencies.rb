module Dapp
  module Build
    module Stage
      # GAArchiveDependencies
      class GAArchiveDependencies < GADependenciesBase
        def initialize(application, next_stage)
          @prev_stage = BeforeInstall.new(application, self)
          super
        end
      end # GAArchiveDependencies
    end # Stage
  end # Build
end # Dapp
