module Dapp
  module Build
    module Stage
      # Artifact
      class Artifact < Base
        include Mod::Artifact

        def initialize(application, next_stage)
          @prev_stage = InstallGroup::GAPostInstallPatch.new(application, self)
          super
        end

        def dependencies
          artifacts_signatures
        end

        def image
          super do |image|
            artifacts_labels = {}
            artifacts.each do |artifact|
              apply_artifact(artifact, image)
              artifacts_labels["dapp-artifact-#{artifact[:name]}".to_sym] = artifact[:app].send(:last_stage).image.built_id
            end
            image.add_service_change_label artifacts_labels
          end
        end

        protected

        def should_not_be_detailed?
          true
        end

        def ignore_log_commands?
          true
        end
      end # Artifact
    end # Stage
  end # Build
end # Dapp
