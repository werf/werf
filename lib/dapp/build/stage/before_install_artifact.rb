module Dapp
  module Build
    module Stage
      # BeforeInstallArtifact
      class BeforeInstallArtifact < ArtifactDefault
        def initialize(application, next_stage)
          @prev_stage = BeforeInstall.new(application, self)
          super
        end
      end # BeforeInstallArtifact
    end # Stage
  end # Build
end # Dapp
