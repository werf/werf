module Dapp
  module Deployment
    module CoreExt
      module Hash
        def kube_in_depth_merge(hash)
          merge(hash) do |_, v1, v2|
            if v1.is_a?(::Hash) && v2.is_a?(::Hash)
              v1.kube_in_depth_merge(v2)
            else
              v2
            end
          end
        end
      end
    end
  end
end

::Hash.send(:include, ::Dapp::Deployment::CoreExt::Hash)
