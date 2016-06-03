module Dapp
  module Stage
    class Source2 < Base
      def signature
        hashsum [builder.stages[:app_install].signature,
                 *builder.infra_setup_commands, # TODO chef
                 *builder.git_artifact_list.map { |git_artifact| git_artifact.source_2_commit }]
      end
    end # Source2
  end # Stage
end # Dapp
