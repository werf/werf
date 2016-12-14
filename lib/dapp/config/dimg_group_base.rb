module Dapp
  module Config
    # DimgGroupBase
    class DimgGroupBase < Base
      attr_reader :_dimg_group

      def initialize(project:)
        @_dimg = []
        @_dimg_group = []

        super(project: project)
      end

      def dev_mode
        @_dev_mode = true
      end

      def dimg(name = nil, &blk)
        Config::Dimg.new(name, project: project).tap do |dimg|
          dimg.instance_eval(&blk) if block_given?
          after_dimg_eval(dimg)
          @_dimg << dimg
        end
      end

      def dimg_group(&blk)
        Config::DimgGroup.new(project: project).tap do |dimg_group|
          dimg_group.instance_eval(&blk) if block_given?
          after_dimg_group_eval(dimg_group)
          @_dimg_group << dimg_group
        end
      end

      def _dimg
        (@_dimg + @_dimg_group.map(&:_dimg)).flatten
      end

      protected

      def after_dimg_eval(dimg)
        pass_to(dimg)
      end

      def after_dimg_group_eval(dimg_group)
        pass_to(dimg_group)
      end

      def pass_to(obj)
        obj.instance_variable_set(:@_dev_mode, obj.instance_variable_get(:@_dev_mode) || @_dev_mode)
      end
    end
  end
end
