module Dapp
  module Kube
    module Kubernetes::Client::Resource
      class Deployment < Base
        def replicas
          spec.fetch('spec', {}).fetch('replicas', nil)
        end
      end # Deployment
    end # Kubernetes::Client::Resource
  end # Kube
end # Dapp
