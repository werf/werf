module Dapp
  class Dapp
    module Command
      module Common
        def option_repo
          unless options[:repo].nil?
            return "localhost:5000/#{name}" if options[:repo] == ':minikube'
            options[:repo]
          end
        end

        def dimg_registry(repo)
          validate_repo_name!(repo)
          ::Dapp::Dimg::DockerRegistry.new(repo)
        end
      end
    end
  end
end
