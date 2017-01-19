module Dapp
  # Artifact
  class Artifact < Dimg
    def after_stages_build!
    end

    def artifact?
      true
    end

    def should_be_built?
      false
    end

    def last_stage
      @last_stage ||= Build::Stage::BuildArtifact.new(self)
    end
  end # Artifact
end # Dapp
