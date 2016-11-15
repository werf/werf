module Dapp
  module Config
    # DimgGroupMain
    class DimgGroupMain < DimgGroupBase
      def dimg(name = nil)
        with_dimg_validation { super }
      end

      def dimg_group
        with_dimg_validation { super }
      end

      protected

      def with_dimg_validation
        yield
        raise Error::Config, code: :dimg_name_required if _dimg.any? { |dimg| dimg._name.nil? } && _dimg.size > 1
      end
    end
  end
end
