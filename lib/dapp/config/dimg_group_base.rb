module Dapp
  module Config
    class DimgGroupBase < Base
      def initialize(project:)
        @_dimgs = []
        @_dimgs_groups = []

        super(project: project)
      end

      def dimg(name, &blk)
        Config::Dimg.new(name, project: _project, &blk).tap do |dimg|
          @_dimgs << dimg
        end
      end

      def dimg_group(&blk)
        Config::DimgGroup.new(project: _project, &blk).tap { |dimg_group| @_dimgs_groups << dimg_group }
      end

      def _dimgs
        (@_dimgs + @_dimgs_groups.map(&:_dimgs)).flatten
      end
    end
  end
end
