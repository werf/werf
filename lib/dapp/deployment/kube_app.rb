module Dapp
  module Deployment
    class KubeApp
      attr_reader :app

      def initialize(app)
        @app = app
      end

      [:deployment, :service].each do |type|
        define_method "#{type}_exist?" do |name|
          public_send("existing_#{type}s_names").include?(name)
        end

        define_method "existing_#{type}s_names" do
          filter_items(app.deployment.kubernetes.public_send(:"#{type}_list")['items'])
        end

        define_method "replace_#{type}!" do |name, spec|
          hash = send(:"merge_kube_#{type}_spec", app.deployment.kubernetes.public_send(type, name), spec)
          app.deployment.kubernetes.public_send(:"replace_#{type}!", name, hash)
        end

        define_method "#{type}_spec_changed?" do |name, spec|
          current_spec = app.deployment.kubernetes.public_send(type, name)
          current_spec != send(:"merge_kube_#{type}_spec", current_spec, spec)
        end

        [:create, :delete].each do |method|
          define_method "#{method}_#{type}!" do |*args|
            app.deployment.kubernetes.public_send(:"#{method}_#{type}!", *args)
          end
        end
      end

      def merge_kube_deployment_spec(spec1, spec2)
        spec1.kube_in_depth_merge(spec2).tap do |spec|
          spec['spec']['template']['spec']['containers'] = begin
            containers1 = spec1['spec']['template']['spec']['containers']
            containers2 = spec2['spec']['template']['spec']['containers']
            containers2.map do |container2|
              if (container1 = containers1.find { |c| c['name'] == container2['name'] }).empty?
                container2
              else
                container1.kube_in_depth_merge(container2)
              end
            end
          end
        end
      end

      def merge_kube_service_spec(spec1, spec2)
        spec1.kube_in_depth_merge(spec2).tap do |spec|
          spec['spec']['ports'] = begin
            ports1 = spec1['spec']['ports']
            ports2 = spec2['spec']['ports']
            ports2.map do |port2|
              if (port1 = ports1.find { |p| p['name'] == port2['name'] }).nil?
                port2
              else
                port1.kube_in_depth_merge(port2)
              end
            end
          end
        end
      end

      protected

      def filter_items(items)
        items.map { |item| item['metadata']['name'] if (app.labels.to_a - item['metadata']['labels'].to_a).empty? }.compact
      end
    end
  end
end
