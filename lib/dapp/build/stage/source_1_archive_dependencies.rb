module Dapp
  module Build
    module Stage
      # Source1ArchiveDependencies
      class Source1ArchiveDependencies < SourceDependenciesBase
        def initialize(application, next_stage)
          @prev_stage = InfraInstall.new(application, self)
          super
        end
      end # Source1ArchiveDependencies
    end # Stage
  end # Build
end # Dapp
