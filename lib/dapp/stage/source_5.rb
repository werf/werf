module Dapp
  module Stage
    class Source5 < Base
      def source_5_actual?
        git_artifact_list.map {|git_artifact| git_artifact.source_5_actual?}.all?
      end

      def source_5_patch
        builder.git_artifact_list.map {|git_artifact| git_artifact.source_5_patch}.reduce(:+)
      end

      def signature
        if source_5_actual?
          builder.stages[:source_4].signature
        else
          hashsum [builder.stages[:source_4].signature, source_5_patch]
        end
      end

      def image
        super do |image|
          builder.git_artifact_list.each do |git_artifact|
            git_artifact.apply_source_5!(image)
          end
        end
      end
    end # Source5
  end # Stage
end # Dapp
