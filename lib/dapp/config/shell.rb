module Dapp
  module Config
    # Shell
    class Shell
      attr_reader :_infra_install, :_infra_setup, :_install, :_setup
      attr_reader :_infra_install_cache_version, :_infra_setup_cache_version, :_install_cache_version, :_setup_cache_version

      def initialize
        @_infra_install = []
        @_infra_setup   = []
        @_install   = []
        @_setup     = []
      end

      def infra_install(*args, cache_version: nil)
        @_infra_install.concat(args)
        @_infra_install_cache_version = cache_version
      end

      def infra_setup(*args, cache_version: nil)
        @_infra_setup.concat(args)
        @_infra_setup_cache_version = cache_version
      end

      def install(*args, cache_version: nil)
        _install.concat(args)
        @_install_cache_version = cache_version
      end

      def setup(*args, cache_version: nil)
        _setup.concat(args)
        @_setup_cache_version = cache_version
      end

      def reset_infra_install
        @_infra_install = []
      end

      def reset_infra_setup
        @_infra_setup = []
      end

      def reset_install
        @_install = []
      end

      def reset_setup
        @_setup = []
      end

      def reset_all
        methods.tap { |arr| arr.delete(__method__) }.grep(/^reset_/).each(&method(:send))
      end

      def clone
        Marshal.load(Marshal.dump(self))
      end
    end
  end
end
