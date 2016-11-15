module Dapp
  module Build
    module Stage
      # AfterInstallArtifact
      class AfterInstallArtifact < ArtifactDefault
        def initialize(dimg, next_stage)
          @prev_stage = InstallGroup::GAPostInstallPatch.new(dimg, self)
          super
        end
      end # AfterInstallArtifact
    end # Stage
  end # Build
end # Dapp
