module Dapp
  module Deployment
    class App
      include Mod::Namespace
      include Mod::SystemEnvironments
      include Mod::Jobs

      attr_reader :deployment
      attr_reader :app_config
      alias config app_config

      def initialize(app_config:, deployment:)
        @app_config = app_config
        @deployment = deployment
      end

      def name(*args)
        deployment.name(app_config._name, *args)
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
              metadata['labels'] = kube.labels
            end
            deployment['spec'] = {}.tap do |spec|
              spec['replicas'] = scale
              spec['template'] = {}
              spec['template']['metadata'] = deployment['metadata']
              spec['template']['spec'] = {}.tap do |template_spec|
                template_spec['containers'] = [].tap do |containers|
                  containers << {}.tap do |container|
                    ports = expose._port.map do |port|
                      {
                        'containerPort' => port._number,
                        'name' => ['app', port._number].join('-'),
                        'protocol' => port._protocol
                      }
                    end
                    volume_mounts = [{'mountPath' => '/.data', 'name' => "volume-#{self.name}"}]

                    container['imagePullPolicy'] = 'Always'
                    container['image']           = [repo, [dimg, image_version].compact.join('-')].join(':')
                    container['name']            = dimg_name
                    container['env']             = environments unless environments.empty?
                    container['ports']           = ports unless expose._port.empty?
                    container['volumeMounts']    = volume_mounts
                  end
                end

                template_spec['volumes'] = [
                  {
                    'name' => "volume-#{name}",
                    'hostPath' => {'path' => "/var/lib/dapp/deployment/volumes/#{self.deployment.name}/#{self.name}"}
                  }
                ]
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
              metadata['labels'] = kube.labels
            end
            service['spec'] = {}.tap do |spec|
              spec['selector'] = kube.labels
              spec['ports']    = expose._port.map do |port|
                {
                  'port' => port._number,
                  'name' => ['service', port._number].join('-'),
                  'protocol' => port._protocol
                }.tap do |h|
                  h['targetPort'] = port._target unless port._target.nil?
                end
              end
              spec['type'] = expose._type
            end
          end
        end
      end

      protected

      def dimg_name
        name(dimg)
      end

      def service_name
        name('service')
      end

      def namespace
        deployment.namespace
      end

      def secret
        deployment.dapp.secret
      end
    end
  end
end
