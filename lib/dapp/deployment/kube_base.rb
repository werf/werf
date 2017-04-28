module Dapp
  module Deployment
    class KubeBase
      def merge_kube_controller_spec(spec1, spec2)
        spec1.kube_in_depth_merge(spec2).tap do |spec|
          spec['spec']['template']['spec']['containers'] = begin
            containers1 = spec1['spec']['template']['spec']['containers']
            containers2 = spec2['spec']['template']['spec']['containers']
            containers2.map do |container2|
              if (container1 = containers1.find { |c| c['name'] == container2['name'] }).nil?
                container2
              else
                container1.kube_in_depth_merge(container2)
              end
            end
          end
        end
      end
    end
  end
end
