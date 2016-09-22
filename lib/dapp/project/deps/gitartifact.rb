module Dapp
  # Project
  class Project
    # Deps
    module Deps
      # Gitartifact
      module Gitartifact
        GITARTIFACT_IMAGE = 'dappdeps/gitartifact:0.1.5'.freeze

        def gitartifact_container_name # FIXME: hashsum(image) or dockersafe()
          GITARTIFACT_IMAGE.tr('/', '_').tr(':', '_')
        end

        def gitartifact_container
          @gitartifact_container ||= begin
            if shellout("docker inspect #{gitartifact_container_name}").exitstatus.nonzero?
              log_secondary_process(t(code: 'process.gitartifact_container_loading'), short: true) do
                shellout!(
                  ['docker create',
                   "--name #{gitartifact_container_name}",
                   "--volume /.dapp/deps/gitartifact #{GITARTIFACT_IMAGE}"].join(' ')
                )
              end
            end
            gitartifact_container_name
          end
        end

        def git_path
          '/.dapp/deps/gitartifact/bin/git'
        end
      end # Gitartifact
    end # Deps
  end # Project
end # Dapp
