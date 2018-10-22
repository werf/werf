module Dapp
  class Dapp
    module GitArtifact
      def dimgstage_g_a_commit_label(paramshash)
        "dapp-git-#{paramshash}-commit"
      end

      def dimgstage_g_a_type_label(paramshash)
        "dapp-git-#{paramshash}-type"
      end
    end # GitArtifact
  end # Dapp
end # Dapp
