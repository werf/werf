module Dapp
  module Dimg
    class Artifact < Dimg
      def name
        dapp.consistent_uniq_slugify(config._name) unless config._name.nil?
      end

      def after_stages_build!
      end

      def stage_should_be_introspected_before_build?(name)
        dapp.options[:introspect_artifact_before] == name
      end

      def stage_should_be_introspected_after_build?(name)
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
