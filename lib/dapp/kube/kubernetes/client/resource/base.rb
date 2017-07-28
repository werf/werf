module Dapp
  module Kube
    module Kubernetes::Client::Resource
      class Base
        attr_reader :spec

        def initialize(spec)
          @spec = spec
        end

        def name
          spec.fetch('metadata', {})['name']
        end

        def annotations
          spec.fetch('metadata', {}).fetch('annotations', {})
        end
      end # Base
    end # Kubernetes::Client::Resource
  end # Kube
end # Dapp
