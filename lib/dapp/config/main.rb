module Dapp
  module Config
    # Application
    class Main < Application
      def initialize(dappfile_path:, project:)
        @project = project

        @_home_path    = Pathname.new(dappfile_path).parent.expand_path.to_s
        @_basename     = Pathname.new(@_home_path).basename.to_s
        @_builder      = Pathname.new(@_home_path).join('Berksfile').exist? ? :chef : :shell
        super(nil)
      end

      def name(value)
        project.log_warning(desc: { code: 'excess_name_instruction', context: 'warning' }) if @_basename == value.to_s
        @_basename = value
      end
    end
  end
end
