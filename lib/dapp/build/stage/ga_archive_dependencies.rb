module Dapp
  module Build
    module Stage
      # GAArchiveDependencies
      class GAArchiveDependencies < GADependenciesBase
        def initialize(dimg, next_stage)
          @prev_stage = BeforeInstallArtifact.new(dimg, self)
          super
        end

        def dependencies
          [dimg.git_artifacts.map(&:paramshash).join]
        end
      end # GAArchiveDependencies
    end # Stage
  end # Build
end # Dapp
