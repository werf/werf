module Dapp
  module Config
    class DimgGroupMain < DimgGroupBase
      def dimg(name = nil)
        with_dimg_validation { super }
      end

      def dimg_group
        with_dimg_validation { super }
      end

      def with_dimg_validation
        yield
        raise if _dimg.any? { |dimg| dimg.instance_variable_get(:@_name).nil? } && _dimg.size > 1 # TODO: only dimg without name
      end
    end
  end
end
