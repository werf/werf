module Dapp
  class Dapp
    module GitArtifact
      def local_git_artifact_exclude_paths(&blk)
        @local_git_artifact_exclude_paths ||= [].tap do |exclude_paths|
          yield exclude_paths if block_given?
        end
      end

      def dimgstage_g_a_commit_label(paramshash)
        "dapp-git-#{paramshash}-commit"
      end

      def dimgstage_g_a_type_label(paramshash)
        "dapp-git-#{paramshash}-type"
      end
    end # GitArtifact
  end # Dapp
end # Dapp
