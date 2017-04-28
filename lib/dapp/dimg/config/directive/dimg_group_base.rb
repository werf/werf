module Dapp
  module Dimg
    module Config
      module Directive
        module DimgGroupBase
          def dimg(name = nil, &blk)
            Dimg.new(name, dapp: dapp).tap do |dimg|
              before_dimg_eval(dimg)
              dimg.instance_eval(&blk) if block_given?
              @_dimg << dimg
            end
          end

          def dimg_group(&blk)
            DimgGroup.new(dapp: dapp).tap do |dimg_group|
              before_dimg_group_eval(dimg_group)
              dimg_group.instance_eval(&blk) if block_given?
              @_dimg_group << dimg_group
            end
          end

          def _dimg
            (@_dimg + @_dimg_group.map(&:_dimg)).flatten
          end

          def _dimg_group
            @_dimg_group
          end

          protected

          def before_dimg_eval(dimg)
          end

          def before_dimg_group_eval(dimg_group)
          end

          def dimg_group_init_variables!
            @_dimg       = []
            @_dimg_group = []
          end
        end
      end
    end
  end
end
