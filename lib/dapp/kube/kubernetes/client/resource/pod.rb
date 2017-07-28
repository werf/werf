module Dapp
  module Kube
    module Kubernetes::Client::Resource
      class Pod < Base
        def container_id(container_name)
          container_status = spec.fetch('status', {})
            .fetch('containerStatuses')
            .find {|cs| cs['name'] == container_name}

          if container_status
            container_status['containerID']
          else
            nil
          end
        end

        def container_state(container_name)
          container_status = spec
            .fetch('status', {})
            .fetch('containerStatuses', [])
            .find {|cs| cs['name'] == container_name}

          if container_status
            container_state, container_state_data = container_status.fetch('state', {}).first
            [container_state, container_state_data]
          else
            [nil, {}]
          end
        end

        def phase
          spec.fetch('status', {}).fetch('phase', nil)
        end

        def containers_names
          spec.fetch('spec', {})
            .fetch('containers', [])
            .map {|container_spec| container_spec['name']}
        end

        def restart_policy
          spec
            .fetch('spec', {})
            .fetch('restartPolicy', nil)
        end
      end # Pod
    end # Kubernetes::Client::Resource
  end # Kube
end # Dapp
