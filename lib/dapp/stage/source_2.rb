module Dapp
  module Stage
    class Source2 < Base
      def name
        :source_2
      end

      def image
        super do |image|
          builder.git_artifact_list.each do |git_artifact|
            git_artifact.apply_source_2!(image)
          end
        end
      end

      def signature
        hashsum [builder.stages[:app_install].signature,
                 *builder.infra_setup_commands, # TODO chef
                 *builder.git_artifact_list.map { |git_artifact| git_artifact.source_2_commit }]
      end
    end # Source2
  end # Stage
end # Dapp
