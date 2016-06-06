module Dapp
  module Stage
    class Source1Archive < Base
      def name
        :source_1_archive
      end

      def image
        super do |image|
          build.git_artifact_list.each do |git_artifact|
            git_artifact.apply_source_1_archive!(image)
          end
        end
      end

      def signature
        hashsum build.stages[:infra_install].signature
      end
    end # Source1Archive
  end # Stage
end # Dapp
