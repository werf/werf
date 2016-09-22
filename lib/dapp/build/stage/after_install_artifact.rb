module Dapp
  module Build
    module Stage
      # AfterInstallArtifact
      class AfterInstallArtifact < ArtifactDefault
        def initialize(application, next_stage)
          @prev_stage = InstallGroup::GAPostInstallPatch.new(application, self)
          super
        end
      end # AfterInstallArtifact
    end # Stage
  end # Build
end # Dapp
