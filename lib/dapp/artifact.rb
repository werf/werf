module Dapp
  # Artifact
  class Artifact < Application
    def initialize(*args)
      super
      @last_stage = Build::Stage::BuildArtifact.new(self)
    end

    def artifact?
      true
    end

    def should_be_built?
      false
    end
  end # Artifact
end # Dapp
