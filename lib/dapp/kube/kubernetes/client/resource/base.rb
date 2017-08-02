module Dapp
  module Kube
    module Kubernetes::Client::Resource
      class Base
        attr_reader :spec

        def initialize(spec)
          @spec = spec
        end

        def metadata
          spec.fetch('metadata', {})
        end

        def name
          metadata['name']
        end

        def uid
          metadata['uid']
        end

        def annotations
          metadata.fetch('annotations', {})
        end

        def status
          spec.fetch('status', {})
        end
      end # Base
    end # Kubernetes::Client::Resource
  end # Kube
end # Dapp
