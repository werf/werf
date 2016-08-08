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

        def dependencies
          artifacts_signatures
        end

        def image
          super do |image|
            artifacts.each { |artifact| apply_artifact(artifact, image) }
          end
        end

        protected

        def should_be_not_detailed?
          true
        end

        def ignore_log_commands?
          true
        end
      end # Artifact
    end # Stage
  end # Build
end # Dapp
