module Dapp
  module Config
    class DimgGroupBase < Base
      attr_reader :_dimg_group

      def initialize(project:)
        @_dimg = []
        @_dimg_group = []

        super(project: project)
      end

      def dimg(name, &blk)
        Config::Dimg.new(name, project: project, &blk).tap do |dimg|
          @_dimg << dimg
        end
      end

      def dimg_group(&blk)
        Config::DimgGroup.new(project: project, &blk).tap { |dimg_group| @_dimg_group << dimg_group }
      end

      def _dimg
        (@_dimg + @_dimg_group.map(&:_dimg)).flatten
      end
    end
  end
end
