module Dapp
  module Kube
    module Kubernetes
      class Config
        extend Helper::YAML
        extend Dapp::Shellout::Base

        class << self
          def new_auto_if_available
            if Kubernetes::Config.kubectl_available?
              Kubernetes::Config.new_from_kubectl
            elsif ENV['KUBECONFIG']
              Kubernetes::Config.new_from_kubeconfig(ENV['KUBECONFIG'])
            else
              default_path = File.join(ENV['HOME'], '.kube/config')
              if File.exists? default_path
                Kubernetes::Config.new_from_kubeconfig(default_path)
              end
            end
          end

          def new_auto
            new_auto_if_available.tap do |cfg|
              raise(Kubernetes::Error::Default,
                code: :config_not_found,
                data: { },
              ) if cfg.nil?
            end
          end

          def new_from_kubeconfig(path)
            unless File.exists?(path)
              raise(Kubernetes::Error::Default,
                code: :config_file_not_found,
                data: { config_path: path }
              )
            end
            self.new yaml_load_file(path), path
          end

          def kubectl_available?
            shellout("kubectl").exitstatus.zero?
          end

          def new_from_kubectl
            cmd_res = shellout(
              "kubectl config view --raw",
              env: {"KUBECONFIG" => ENV["KUBECONFIG"]}
            )

            shellout_cmd_should_succeed! cmd_res

            self.new YAML.load(cmd_res.stdout), "kubectl config view --raw"
          end
        end # << self

        attr_reader :config_hash, :config_path

        def initialize(config_hash, config_path)
          @config_hash = config_hash
          @config_path = config_path
        end

        def context_names
          config_hash.fetch('contexts', []).map {|context| context['name']}
        end

        def current_context_name
          @current_context_name ||= begin
            config_hash['current-context'] || begin
              if (context = config_hash.fetch('contexts', []).first)
                warn "[WARN] .kube/config current-context is not set, using first context '#{context['name']}'"
                context['name']
              end
            end
          end
        end

        def context_config(context_name)
          res = config_hash.fetch('contexts', [])
            .find {|context| context['name'] == context_name}

          raise(Kubernetes::Error::Default,
            code: :context_config_not_found,
            data: {config_path: config_path,
                    config: config_hash,
                    context_name: context_name}
          ) if res.nil?

          res['context']
        end

        def user_config(user_name)
          res = config_hash.fetch('users', [])
            .find {|user| user['name'] == user_name}

          raise(Kubernetes::Error::Default,
            code: :user_config_not_found,
            data: {config_path: config_path,
                    user: user_name}
          ) if res.nil?

          res['user']
        end

        def cluster_config(cluster_name)
          res = config_hash.fetch('clusters', [])
            .find {|cluster| cluster['name'] == cluster_name}

          raise(Kubernetes::Error::Default,
            code: :cluster_config_not_found,
            data: {config_path: config_path,
                  cluster: cluster_name}
          ) if res.nil?

          res['cluster']
        end

        def cluster_name(context_name)
          cfg = context_config(context_name)
          cfg['cluster'] if cfg
        end

        def namespace(context_name)
          cfg = context_config(context_name)
          cfg['namespace'] if cfg
        end

      end # Kubeconfig
    end # Kubernetes
  end # Kube
end # Dapp
