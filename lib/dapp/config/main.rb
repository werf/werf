module Dapp
  module Config
    # Main
    class Main < Application
      def initialize(**options)
        @_home_path    = Pathname.new(options[:dappfile_path]).parent.expand_path.to_s
        @_basename     = Pathname.new(@_home_path).basename.to_s
        @_builder      = Pathname.new(@_home_path).join('Berksfile').exist? ? :chef : :shell

        @_docker       = Docker.new
        @_git_artifact = GitArtifact.new
        @_shell        = Shell.new
        @_chef         = Chef.new

        super(nil)
      end

      def name(value)
        @_basename = value
      end
    end
  end
end
