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
          before_dimg_eval(dimg)
          dimg.instance_eval(&blk) if block_given?
          @_dimg << dimg
        end
      end

      def dimg_group(&blk)
        Config::DimgGroup.new(project: project, &blk).tap { |dimg_group| @_dimg_group << dimg_group }
      end

      def _dimg
        (@_dimg + @_dimg_group.map(&:_dimg)).flatten
      end

      protected

      def before_dimg_eval(dimg)
        dimg.instance_variable_set(:@_dev_mode, @_dev_mode)
      end
    end
  end
end
