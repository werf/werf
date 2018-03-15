module Dapp
  class Dapp
    module Command
      module Common
        def option_repo
          shortcut_or_key(options[:repo])
        end

        def shortcut_or_key(key)
          shortcuts[key] || key
        end

        def shortcuts
          { ':minikube' => "localhost:5000/#{name}" }
        end

        def dimg_registry(repo = option_repo)
          validate_repo_name!(repo)
          ::Dapp::Dimg::DockerRegistry.new(repo)
        end
      end
    end
  end
end
