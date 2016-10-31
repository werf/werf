module Dapp
  module Config
    class DimgGroupMain < DimgGroupBase
      def initialize(dappfile_path:, project:)
        @_home_path    = Pathname.new(dappfile_path).parent.expand_path.to_s
        # @_builder      = Pathname.new(@_home_path).join('Berksfile').exist? ? :chef : :shell

        super(project: project)
      end

      def dimg(name = nil)
        with_dimg_validation { super }
      end

      def dimg_group
        with_dimg_validation { super }
      end

      def with_dimg_validation
        yield
        raise if _dimgs.any? { |dimg| dimg.instance_variable_get(:@_name).nil? } && _dimgs.size > 1 # TODO: only dimg without name
      end
    end
  end
end
