module Dapp
  module Deployment
    class App
      include Namespace
      include SystemEnvironments

      attr_reader :app_config
      attr_reader :deployment

      def initialize(app_config:, deployment:)
        @app_config = app_config
        @deployment = deployment
      end

      def name
        [deployment.dapp.name, app_config._name].compact.join('-').gsub('_', '-')
      end

      def labels
        { 'dapp-app' => name }
      end

      def kube
        @kube ||= KubeApp.new(self)
      end

      [:dimg, :expose, :bootstrap, :migrate].each do |directive|
        define_method directive do
          app_config.public_send("_#{directive}")
        end
      end

      def to_kube_deployments(repo, image_version)
        {}.tap do |hash|
          hash[name] = {}.tap do |deployment|
            deployment['metadata'] = {}.tap do |metadata|
              metadata['name']   = name
              metadata['labels'] = labels
            end
            deployment['spec'] = {}.tap do |spec|
              spec['replicas'] = scale
              spec['template'] = {}
              spec['template']['metadata'] = deployment['metadata']
              spec['template']['spec'] = {}.tap do |template_spec|
                template_spec['containers'] = [].tap do |containers|
                  containers << {}.tap do |container|
                    envs = [environment, secret_environment]
                             .select { |env| !env.empty? }
                             .map { |h| h.map { |k, v| { name: k, value: v } } }
                             .flatten
                    container['env']             = envs unless envs.empty?
                    container['imagePullPolicy'] = 'Always'
                    container['image']           = [repo, [dimg, image_version].compact.join('-')].join(':')
                    container['name']            = dimg_name
                    container['ports']           = expose._port.map do |p|
                      p._list.each_with_index.map do |port, ind|
                        { "containerPort" => port, 'name' => ['app', ind].join('-'), "protocol" => p._protocol }
                      end
                    end.flatten
                  end
                end
              end
            end
          end
        end
      end

      def to_kube_services
        return {} if expose._port.empty?

        {}.tap do |hash|
          hash[service_name] = {}.tap do |service|
            service['metadata'] = {}.tap do |metadata|
              metadata['name']   = service_name
              metadata['labels'] = labels
            end
            service['spec'] = {}.tap do |spec|
              spec['selector'] = labels
              spec['ports']    = expose._port.map do |p|
                p._list.each_with_index.map do |port, ind|
                  { 'port' => port, 'name' => ['service', ind].join('-'), 'protocol' => p._protocol }
                end
              end.flatten
            end
          end
        end
      end

      protected

      def dimg_name
        [name, dimg].compact.join('-').gsub('_', '-')
      end

      def service_name
        [name, 'service'].join('-')
      end
    end
  end
end
