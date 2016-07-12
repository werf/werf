module Dapp
  module Config
    class Main < Application
      def initialize(**options)
        @_home_path = Pathname.new(options[:dappfile_path]).parent.expand_path.to_s
        @_name      = Pathname.new(@_home_path).basename.to_s
        @_builder   = Pathname.new(@_home_path).join('Berksfile').exist? ? :chef : :shell
        super(nil)
      end

      def name(value)
        @_name = value
      end
    end
  end
end
