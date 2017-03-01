module Dapp
  module Dimg
    module Config
      # DimgGroupMain
      class DimgGroupMain < DimgGroupBase
        def dimg(name = nil, &blk)
          with_dimg_validation { super(name, &blk) }
        end

        def dimg_group(&blk)
          with_dimg_validation { super(&blk) }
        end

        protected

        def with_dimg_validation
          yield
          raise Error::Config, code: :dimg_name_required if _dimg.any? { |dimg| dimg._name.nil? } && _dimg.size > 1
        end
      end
    end
  end
end
