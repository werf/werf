module Dapp
  class Dapp
    module Deps
      module Gitartifact
        GITARTIFACT_VERSION = "0.2.1"

        def gitartifact_container_name # FIXME: hashsum(image) or dockersafe()
          "dappdeps_gitartifact_#{GITARTIFACT_VERSION}"
        end

        def gitartifact_container
          @gitartifact_container ||= begin
            is_container_exist = proc{shellout("#{host_docker} inspect #{gitartifact_container_name}").exitstatus.zero?}
            if !is_container_exist.call
              lock("dappdeps.container.#{gitartifact_container_name}", default_timeout: 120) do
                if !is_container_exist.call
                  log_secondary_process(t(code: 'process.gitartifact_container_creating', data: {name: gitartifact_container_name}), short: true) do
                    shellout!(
                      ["#{host_docker} create",
                      "--name #{gitartifact_container_name}",
                      "--volume /.dapp/deps/gitartifact/#{GITARTIFACT_VERSION}",
                      "dappdeps/gitartifact:#{GITARTIFACT_VERSION}"].join(' ')
                    )
                  end
                end
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
  end # Dapp
end # Dapp
