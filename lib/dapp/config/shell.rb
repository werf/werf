module Dapp
  module Config
    class Shell < Base
      attr_accessor :infra_install, :infra_setup, :app_install, :app_setup

      def initialize(main_conf, &blk)
        main_conf.builder_validation(:shell)
        @infra_install = []
        @infra_setup   = []
        @app_install   = []
        @app_setup     = []
        super
      end
    end
  end
end
