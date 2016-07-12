module Dapp
  module Config
    class Application < Base
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
        super()
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
        instance_variable_set(:"@_#{another_builder}", nil)
        @_builder = type
      end

      def _apps
        @_apps.empty? ? [self] : @_apps.flatten
      end

      def to_h
        {
            name:         _name,
            builder:      _builder,
            docker:       @_docker.to_h,
            git_artifact: @_git_artifact.to_h,
            shell:        @_shell.to_h,
            chef:         @_chef.to_h
        }
      end

      private

      def clone
        Application.new(self).tap do |app|
          app.instance_variable_set(:'@_builder', _builder)
          app.instance_variable_set(:'@_home_path', _home_path)
          app.instance_variable_set(:'@_docker', _docker.clone_with_marshal)             unless @_docker.nil?
          app.instance_variable_set(:'@_git_artifact', _git_artifact.clone_with_marshal) unless @_git_artifact.nil?
          app.instance_variable_set(:'@_chef', _chef.clone_with_marshal)                 unless @_chef.nil?
          app.instance_variable_set(:'@_shell', _shell.clone_with_marshal)               unless @_shell.nil?
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
