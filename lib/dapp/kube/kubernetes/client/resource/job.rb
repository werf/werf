module Dapp
  module Kube
    module Kubernetes::Client::Resource
      class Job < Base
        def uid
          spec.fetch('metadata', {})['uid']
        end

        def terminated?
          failed? || succeeded?
        end

        def failed?
          !!spec.fetch('status', {}).fetch('conditions', []).find do |cond|
            cond['type'] == 'Failed'
          end
        end

        def succeeded?
          !!spec.fetch('status', {}).fetch('conditions', []).find do |cond|
            cond['type'] == 'Complete'
          end
        end
      end # Job
    end # Kubernetes::Client::Resource
  end # Kube
end # Dapp
