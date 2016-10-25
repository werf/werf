module Dapp
  # Artifact
  class Artifact < Application
    def artifact?
      true
    end

    def should_be_built?
      false
    end

    protected

    def last_stage
      @last_stage ||= Build::Stage::BuildArtifact.new(self)
    end
  end # Artifact
end # Dapp
