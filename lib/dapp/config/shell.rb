module Dapp
  module Config
    class Shell < Base
      attr_accessor :_infra_install, :_infra_setup, :_app_install, :_app_setup

      def initialize
        @_infra_install = []
        @_infra_setup   = []
        @_app_install   = []
        @_app_setup     = []
        super
      end

      def infra_install(*args, cache_version: nil)
        @_infra_install.push(*args.flatten)
        cache_version(infra_install: cache_version) unless cache_version.nil?
      end

      def infra_setup(*args, cache_version: nil)
        @_infra_setup.push(*args.flatten)
        cache_version(infra_setup: cache_version) unless cache_version.nil?
      end

      def app_install(*args, cache_version: nil)
        @_app_install.push(*args.flatten)
        cache_version(app_install: cache_version) unless cache_version.nil?
      end

      def app_setup(*args, cache_version: nil)
        @_app_setup.push(*args.flatten)
        cache_version(app_setup: cache_version) unless cache_version.nil?
      end

      def to_h
        {
          infra_install: _infra_install,
          infra_setup:   _infra_setup,
          app_install:   _app_install,
          app_setup:     _app_setup
        }
      end
    end
  end
end
