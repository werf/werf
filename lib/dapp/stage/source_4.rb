module Dapp
  module Stage
    class Source4 < Base
      def name
        :source_4
      end

      def source_4_patch
        builder.git_artifact_list.map {|git_artifact| git_artifact.source_4_patch}.reduce(:+)
      end

      def source_4_actual?
        builder.git_artifact_list.map {|git_artifact| git_artifact.source_4_actual?}.all?
      end

      def source_4_commit_list
        builder.git_artifact_list.map {|git_artifact| git_artifact.source_4_commit}
      end

      def signature
        if source_4_actual? or source_4_patch.bytesize < 50*1024*1024
          builder.stages[:app_setup].signature
        else
          hashsum [builder.stages[:app_setup].signature, *source_4_commit_list]
        end
      end

      def image
        super do |image|
          builder.git_artifact_list.each do |git_artifact|
            git_artifact.apply_source_4!(image)
          end
        end
      end
    end # Source4
  end # Stage
end # Dapp
