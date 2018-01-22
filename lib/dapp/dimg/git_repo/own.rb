module Dapp
  module Dimg
    module GitRepo
      class Own < Local
        def initialize(dapp)
          super(dapp, 'own', dapp.path.to_s)
        end

        def exclude_paths
          dapp.local_git_artifact_exclude_paths
        end
      end
    end
  end
end
