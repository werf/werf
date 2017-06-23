module Dapp
  class Dapp
    module Deps
      module Gitartifact
        GITARTIFACT_VERSION = '0.1.7'.freeze

        def gitartifact_container_name # FIXME: hashsum(image) or dockersafe()
          "dappdeps_gitartifact_#{GITARTIFACT_VERSION}"
        end

        def gitartifact_container
          @gitartifact_container ||= begin
            unless docker_client.container?(gitartifact_container_name)
              log_secondary_process(t(code: 'process.gitartifact_container_creating')) do
                with_log_indent do
                  hostconfig = {}
                  hostconfig[:mounts] = [{ target: "/.dapp/deps/gitartifact/#{GITARTIFACT_VERSION}", type: :volume }]
                  volumes = { "/.dapp/deps/gitartifact/#{GITARTIFACT_VERSION}" => {} }
                  docker_client.container_create(name: gitartifact_container_name,
                                                 image: "dappdeps/gitartifact:#{GITARTIFACT_VERSION}",
                                                 volumes: volumes,
                                                 hostconfig: hostconfig)
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
