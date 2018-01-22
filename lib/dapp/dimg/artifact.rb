module Dapp
  module Dimg
    class Artifact < Dimg
      def after_stages_build!
      end

      def stage_should_be_introspected?(name)
        dapp.options[:introspect_artifact_stage] == name
      end

      def artifact?
        true
      end

      def should_be_built?
        false
      end

      def last_stage_class
        Build::Stage::BuildArtifact
      end
    end # Artifact
  end # Dimg
end # Dapp
