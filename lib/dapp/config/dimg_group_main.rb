module Dapp
  module Config
    class DimgGroupMain < DimgGroupBase
      def initialize(dappfile_path:, project:)
        @_home_path    = Pathname.new(dappfile_path).parent.expand_path.to_s
        # @_builder      = Pathname.new(@_home_path).join('Berksfile').exist? ? :chef : :shell

        super(project: project, basename: Pathname.new(@_home_path).basename.to_s)
      end
    end
  end
end
