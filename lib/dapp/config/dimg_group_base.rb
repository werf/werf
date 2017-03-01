module Dapp
  module Config
    # DimgGroupBase
    class DimgGroupBase < Base
      attr_reader :_dimg_group

      def initialize(dapp:)
        @_dimg = []
        @_dimg_group = []

        super(dapp: dapp)
      end

      def dev_mode
        @_dev_mode = true
      end

      def dimg(name = nil, &blk)
        Config::Dimg.new(name, dapp: dapp).tap do |dimg|
          before_dimg_eval(dimg)
          dimg.instance_eval(&blk) if block_given?
          @_dimg << dimg
        end
      end

      def dimg_group(&blk)
        Config::DimgGroup.new(dapp: dapp).tap do |dimg_group|
          before_dimg_group_eval(dimg_group)
          dimg_group.instance_eval(&blk) if block_given?
          @_dimg_group << dimg_group
        end
      end

      def _dimg
        (@_dimg + @_dimg_group.map(&:_dimg)).flatten
      end

      protected

      def before_dimg_eval(dimg)
        pass_to_default(dimg)
      end

      def before_dimg_group_eval(dimg_group)
        pass_to_default(dimg_group)
      end

      def pass_to_default(obj)
        obj.instance_variable_set(:@_dev_mode, @_dev_mode)
      end
    end
  end
end
