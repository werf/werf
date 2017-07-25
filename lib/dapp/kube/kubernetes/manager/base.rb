module Dapp
  module Kube
    module Kubernetes::Manager
      class Base
        attr_reader :dapp
        attr_reader :name

        def initialize(dapp, name)
          @dapp = dapp
          @name = name
        end
      end # Base
    end # Kubernetes::Manager
  end # Kube
end # Dapp
