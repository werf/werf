module Dapp
  module Config
    class DimgGroupBase < Base
      attr_reader :_basename

      def initialize(project:, basename:)
        @_dimgs = []
        @_dimgs_groups = []
        @_basename = basename

        super(project: project)
      end

      def dimg(name = nil, &blk)
        raise if name.nil? && _dimgs.size >= 1 # TODO: only dimg without name
        Config::Dimg.new(dimg_name(name), project: _project, &blk).tap do |dimg|
          @_dimgs << dimg
        end
      end

      def dimg_name(name)
        File.join(*[_basename, name].compact)
      end

      def dimg_group(&blk)
        Config::DimgGroup.new(project: _project, basename: _basename, &blk).tap { |dimg_group| @_dimgs_groups << dimg_group }
      end

      def _dimgs
        (@_dimgs + @_dimgs_groups.map(&:_dimgs)).flatten
      end
    end
  end
end
