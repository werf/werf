module Dapp
  module Stage
    class Source2 < Base
      def name
        :source_2
      end

      def prev_source_stage_name
        :source_1
      end

      def image
        super do |image|
          build.git_artifact_list.each do |git_artifact|
            git_artifact.source_2_apply!(image)
          end
        end
      end

      def signature
        hashsum [build.stages[:app_install].signature,
                 *build.infra_setup_commands, # TODO chef
                 *build.git_artifact_list.map { |git_artifact| git_artifact.source_2_commit }]
      end

      def git_artifact_signature
        hashsum [build.stages[:app_install].signature,
                 *build.infra_setup_commands]
      end
    end # Source2
  end # Stage
end # Dapp
