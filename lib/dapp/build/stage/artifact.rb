module Dapp
  module Build
    module Stage
      # Artifact
      class Artifact < Base
        include Mod::Artifact

        def initialize(application, next_stage)
          @prev_stage = Install.new(application, self)
          super
        end

        def signature
          hashsum [super, *artifacts_signatures]
        end

        def image
          super do |image|
            artifacts.each { |artifact| apply_artifact(artifact, image) }
          end
        end
      end # Artifact
    end # Stage
  end # Build
end # Dapp
