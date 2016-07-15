module Dapp
  module Config
    # Application
    class Application
      attr_reader :_name
      attr_reader :_builder
      attr_reader :_home_path
      attr_reader :_docker
      attr_reader :_git_artifact
      attr_reader :_chef
      attr_reader :_shell
      attr_reader :_parent

      def initialize(parent)
        @_apps      = []
        @_parent    = parent

        yield self if block_given?
      end

      def chef
        raise 'Already defined another builder type!' unless _builder == :chef
        @_chef ||= Chef.new
      end

      def shell
        raise 'Already defined another builder type!' unless _builder == :shell
        @_shell ||= Shell.new
      end

      def git_artifact
        @_git_artifact ||= GitArtifact.new
      end

      def docker
        @_docker ||= Docker.new
      end

      def builder(type)
        raise "Builder type `#{type}` isn't supported!" unless [:chef, :shell].include?((type = type.to_sym))
        another_builder = [:chef, :shell].find { |t| t != type }
        instance_variable_set(:"@_#{another_builder}", Config.const_get(another_builder.capitalize).new)
        @_builder = type
      end

      def _apps
        @_apps.empty? ? [self] : @_apps.flatten
      end

      def to_h
        compact(name:         _name,
                builder:      _builder,
                docker:       @_docker.to_h,
                git_artifact: @_git_artifact.to_h,
                shell:        @_shell.to_h,
                chef:         @_chef.to_h)
      end

      def _app_runlist
        @_app_runlist ||= (_parent ? _parent._app_runlist : []) + [self]
      end

      def _root_app
        _app_runlist.first
      end

      private

      def clone
        Application.new(self).tap do |app|
          app.instance_variable_set(:'@_builder', _builder)
          app.instance_variable_set(:'@_home_path', _home_path)
          app.instance_variable_set(:'@_docker', _docker.clone)             unless @_docker.nil?
          app.instance_variable_set(:'@_git_artifact', _git_artifact.clone) unless @_git_artifact.nil?
          app.instance_variable_set(:'@_chef', _chef.clone)                 unless @_chef.nil?
          app.instance_variable_set(:'@_shell', _shell.clone)               unless @_shell.nil?
        end
      end

      def app(sub_name, &blk)
        clone.tap do |app|
          app.instance_variable_set(:'@_name', [_name, sub_name].compact.join('-'))
          app.instance_eval(&blk) if block_given?
          @_apps += app._apps
        end
      end

      def compact(hash)
        hash.delete_if do |_key, val|
          case val
          when Hash   then compact(val).empty?
          when Array  then val.map { |v| v.is_a?(Hash) ? compact(v) : v }.empty?
          when String then val.empty?
          else val.nil?
          end
        end
      end
    end
  end
end
