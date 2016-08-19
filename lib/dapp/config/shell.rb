module Dapp
  module Config
    # Shell
    class Shell
      attr_reader :_before_install, :_before_setup, :_install, :_setup
      attr_reader :_before_install_cache_version, :_before_setup_cache_version, :_install_cache_version, :_setup_cache_version

      def initialize
        @_before_install = []
        @_before_setup   = []
        @_install       = []
        @_setup         = []
      end

      def before_install(*args, cache_version: nil)
        @_before_install.concat(args)
        @_before_install_cache_version = cache_version
      end

      def before_setup(*args, cache_version: nil)
        @_before_setup.concat(args)
        @_before_setup_cache_version = cache_version
      end

      def install(*args, cache_version: nil)
        _install.concat(args)
        @_install_cache_version = cache_version
      end

      def setup(*args, cache_version: nil)
        _setup.concat(args)
        @_setup_cache_version = cache_version
      end

      def reset_before_install
        @_before_install = []
      end

      def reset_before_setup
        @_before_setup = []
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

      def empty?
        @_before_install.empty? && @_before_setup.empty? && @_install.empty? && @_setup.empty?
      end

      def clone
        Marshal.load(Marshal.dump(self))
      end
    end
  end
end
