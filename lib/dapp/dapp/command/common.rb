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

        def dimg_name!
          one_dimg!
          build_configs.first._name
        end

        def one_dimg!
          return if build_configs.one?
          raise Error::Command, code: :command_unexpected_dimgs_number, data: { dimgs_names: build_configs.map(&:_name).join(' ') }
        end

        def dimg_registry(repo)
          validate_repo_name!(repo)
          ::Dapp::Dimg::DockerRegistry.new(repo)
        end
      end
    end
  end
end
