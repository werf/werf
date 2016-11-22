module Dapp
  # Project
  class Project
    # Deps
    module Deps
      # Gitartifact
      module Gitartifact
        GITARTIFACT_VERSION = '0.1.7'.freeze

        def gitartifact_container_name # FIXME: hashsum(image) or dockersafe()
          "dappdeps_gitartifact_#{GITARTIFACT_VERSION}"
        end

        def gitartifact_container
          @gitartifact_container ||= begin
            if shellout("docker inspect #{gitartifact_container_name}").exitstatus.nonzero?
              log_secondary_process(t(code: 'process.gitartifact_container_creating'), short: true) do
                shellout!(
                  ['docker create',
                   "--name #{gitartifact_container_name}",
                   "--volume /.dapp/deps/gitartifact/#{GITARTIFACT_VERSION}",
                   "dappdeps/gitartifact:#{GITARTIFACT_VERSION}"].join(' ')
                )
              end
            end
            gitartifact_container_name
          end
        end

        def git_bin
          "/.dapp/deps/gitartifact/#{GITARTIFACT_VERSION}/bin/git"
        end
      end # Gitartifact
    end # Deps
  end # Project
end # Dapp
