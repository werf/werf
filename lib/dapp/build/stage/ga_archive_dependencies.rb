module Dapp
  module Build
    module Stage
      # GAArchiveDependencies
      class GAArchiveDependencies < GADependenciesBase
        def initialize(application, next_stage)
          @prev_stage = BeforeInstallArtifact.new(application, self)
          super
        end

        def dependencies
          [application.git_artifacts.map(&:paramshash).join]
        end
      end # GAArchiveDependencies
    end # Stage
  end # Build
end # Dapp
