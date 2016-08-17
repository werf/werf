module Dapp
  module Config
    # Application
    class Application
      attr_reader :_builder
      attr_reader :_home_path
      attr_reader :_basename
      attr_reader :_docker
      attr_reader :_git_artifact
      attr_reader :_artifact
      attr_reader :_chef
      attr_reader :_shell
      attr_reader :_parent
      attr_reader :_install_dependencies
      attr_reader :_setup_dependencies
      attr_reader :_parent

      def initialize(parent)
        @_apps   = []
        @_parent = parent

        @_artifact = []
        @_install_dependencies = []
        @_setup_dependencies   = []

        yield self if block_given?
      end

      def install_depends_on(*args)
        @_install_dependencies.concat(args)
      end

      def setup_depends_on(*args)
        @_setup_dependencies.concat(args)
      end

      def builder(type)
        controller.log_warning(desc: { code: 'excess_builder_instruction', context: 'warning' }) if @_chef.empty? && @_shell.empty?
        raise Error::Config, code: :builder_type_unsupported, data: { type: type } unless [:chef, :shell].include?((type = type.to_sym))
        another_builder = [:chef, :shell].find { |t| t != type }
        instance_variable_set(:"@_#{another_builder}", Config.const_get(another_builder.capitalize).new)
        @_builder = type
      end

      def chef
        raise Error::Config, code: :builder_type_conflict unless _builder == :chef
        @_chef ||= Chef.new
      end

      def shell
        raise Error::Config, code: :builder_type_conflict unless _builder == :shell
        @_shell ||= Shell.new
      end

      def docker
        @_docker ||= Docker.new
      end

      def artifact(where_to_add, **options, &blk)
        @_artifact << begin
          config = clone.tap do |app|
            app.instance_variable_set(:'@_artifact', [])
            app.instance_variable_set(:'@_name', app_name("artifact-#{SecureRandom.hex(2)}"))
            app.instance_eval(&blk) if block_given?
          end
          Artifact::Stage.new(where_to_add, config: config, **options)
        end
      end

      def git_artifact
        @_git_artifact ||= GitArtifact.new
      end

      def _name
        (@_name || @_basename).tap do |name|
          reg = '^[[[:alnum:]]_.-]*$'
          raise Error::Config, code: :app_name_incorrect, data: { name: name, reg: reg } unless name =~ /#{reg}/
        end
      end

      def _apps
        @_apps.empty? ? [self] : @_apps.flatten
      end

      def _app_chain
        @_app_chain ||= (_parent ? _parent._app_chain : []) + [self]
      end

      def _app_runlist
        _app_chain.map(&:_name).map do |name|
          if (subname = name.split("#{_root_app._name}-", 2)[1])
            subname_parts = subname.split('-')
            subname_parts.join('_') if subname_parts.any?
          end
        end.compact
      end

      def _root_app
        _app_chain.first
      end

      protected

      attr_accessor :controller

      private

      # rubocop:disable Metrics/AbcSize
      def clone
        Application.new(self).tap do |app|
          app.instance_variable_set(:'@controller', controller)
          app.instance_variable_set(:'@_builder', _builder)
          app.instance_variable_set(:'@_home_path', _home_path)
          app.instance_variable_set(:'@_basename', _basename)
          app.instance_variable_set(:'@_install_dependencies', _install_dependencies)
          app.instance_variable_set(:'@_setup_dependencies', _setup_dependencies)
          app.instance_variable_set(:'@_artifact', Marshal.load(Marshal.dump(_artifact)))
          app.instance_variable_set(:'@_docker', _docker.clone)             unless @_docker.nil?
          app.instance_variable_set(:'@_git_artifact', _git_artifact.clone) unless @_git_artifact.nil?
          app.instance_variable_set(:'@_chef', _chef.clone)                 unless @_chef.nil?
          app.instance_variable_set(:'@_shell', _shell.clone)               unless @_shell.nil?
        end
      end
      # rubocop:enable Metrics/AbcSize

      def app(sub_name, &blk)
        clone.tap do |app|
          app.instance_variable_set(:'@_name', app_name(sub_name))
          app.instance_eval(&blk) if block_given?
          @_apps += app._apps
        end
      end

      def app_name(sub_name)
        [_name, sub_name].compact.join('-')
      end
    end
  end
end
