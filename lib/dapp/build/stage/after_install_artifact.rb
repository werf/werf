module Dapp
  module Build
    module Stage
      # AfterInstallArtifact
      class AfterInstallArtifact < ArtifactBase
        def initialize(application, next_stage)
          @prev_stage = InstallGroup::GAPostInstallPatch.new(application, self)
          super
        end
      end # AfterInstallArtifact
    end # Stage
  end # Build
end # Dapp
