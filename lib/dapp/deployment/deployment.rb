module Dapp
  module Deployment
    class Deployment
      include Mod::Namespace
      include Mod::SystemEnvironments

      attr_reader :dapp

      def initialize(dapp:)
        @dapp = dapp
      end

      def name(*args)
        [dapp.name, *args].flatten.compact.join('-').gsub('_', '-')
      end

      def kube
        @kube ||= KubeDeployment.new(self)
      end

      def apps
        @apps ||= dapp.apps_configs.map { |app_config| App.new(app_config: app_config, deployment: self) }
      end

      def namespace
        dapp.options[:namespace] || ENV['DAPP_NAMESPACE']
      end

      def kubernetes
        @kubernetes ||= Kubernetes.new(namespace: namespace)
      end

      def to_kube_bootstrap_pods(repo, image_version)
        return {} if deployment_config._bootstrap.empty?

        {}.tap do |hash|
          hash[bootstrap_pod_name] = {}.tap do |pod|
            pod['metadata'] = {}.tap do |metadata|
              metadata['name']   = bootstrap_pod_name
              metadata['labels'] = kube.labels
            end
            pod['spec'] = {}.tap do |spec|
              spec['restartPolicy'] = 'Never'
              spec['containers']    = [].tap do |containers|
                containers << {}.tap do |container|
                  envs = [environment, secret_environment]
                           .select { |env| !env.empty? }
                           .map { |h| h.map { |k, v| { name: k, value: v } } }
                           .flatten
                  container['env']             = envs unless envs.empty?
                  container['imagePullPolicy'] = 'Always'
                  container['command']         = deployment_config._bootstrap._run
                  container['image']           = [repo, [dapp.config._bootstrap._dimg, image_version].compact.join('-')].join(':')
                  container['name']            = bootstrap_pod_name
                end
              end
            end
          end
        end
      end

      protected

      def bootstrap_pod_name
        name('bootstrap')
      end

      def deployment_config
        dapp.config._deployment
      end
      alias config deployment_config

      def secret
        dapp.secret
      end
    end
  end
end
