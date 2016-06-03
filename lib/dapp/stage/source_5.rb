module Dapp
  module Stage
    class Source5 < Base
      def signature
        if git_artifact_list.map {|git_artifact| git_artifact.source_5_actual?}.all?
          builder.stages[:source_4].signature
        else
          hashsum [builder.stages[:source_4].signature, git_artifact_patch]
        end
      end
    end # Source5
  end # Stage
end # Dapp
