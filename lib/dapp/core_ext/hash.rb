module Dapp
  module CoreExt
    module Hash
      def in_depth_merge(hash) # do not conflict with activesupport`s deep_merge
        merge(hash) do |_, v1, v2|
          if v1.is_a?(::Hash) && v2.is_a?(::Hash)
            v1.in_depth_merge(v2)
          elsif v1.is_a?(::Array) || v2.is_a?(::Array)
            [v1, v2].flatten
          else
            v2
          end
        end
      end
    end
  end
end

::Hash.send(:include, ::Dapp::CoreExt::Hash)
