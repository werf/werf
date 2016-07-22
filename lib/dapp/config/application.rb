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
      attr_reader :_app_install_dependencies
      attr_reader :_app_setup_dependencies
      attr_reader :_parent

      def initialize(parent)
        @_apps   = []
        @_parent = parent

        @_app_install_dependencies = []
        @_app_setup_dependencies   = []

        yield self if block_given?
      end

      def app_install_depends_on(*args)
        @_app_install_dependencies.concat(args)
      end

      def app_setup_depends_on(*args)
        @_app_setup_dependencies.concat(args)
      end

      def chef
        fail Error::Config, code: :another_builder_defined unless _builder == :chef
        @_chef ||= Chef.new
      end

      def shell
        fail Error::Config, code: :another_builder_defined unless _builder == :shell
        @_shell ||= Shell.new
      end

      def git_artifact
        @_git_artifact ||= GitArtifact.new
      end

      def docker
        @_docker ||= Docker.new
      end

      def builder(type)
        fail Error::Config, code: :builder_type_is_not_supported, data: { type: type } unless [:chef, :shell].include?((type = type.to_sym))
        another_builder = [:chef, :shell].find { |t| t != type }
        instance_variable_set(:"@_#{another_builder}", Config.const_get(another_builder.capitalize).new)
        @_builder = type
      end

      def _apps
        @_apps.empty? ? [self] : @_apps.flatten
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
          app.instance_variable_set(:'@_app_install_dependencies', _app_install_dependencies)
          app.instance_variable_set(:'@_app_setup_dependencies', _app_setup_dependencies)
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
    end
  end
end
