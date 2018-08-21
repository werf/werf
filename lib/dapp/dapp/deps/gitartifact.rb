module Dapp
  class Dapp
    module Deps
      module Gitartifact
        def gitartifact_container
          dappdeps_container(:gitartifact)
        end

        def git_bin
          ruby2go_dappdeps_command(dappdeps: :gitartifact, command: :bin)
        end
      end # Gitartifact
    end # Deps
  end # Dapp
end # Dapp
