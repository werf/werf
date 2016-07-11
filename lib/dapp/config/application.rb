module Dapp
  module Config
    class Application < Base
      attr_reader :_name
      attr_reader :_builder
      attr_reader :_home_path
      attr_reader :apps

      def initialize(**options)
        unless options.empty?
          @_home_path = Pathname.new(options[:dappfile_path]).parent.expand_path.to_s
          @_name      = Pathname.new(@_home_path).basename
        end
        @apps      = []
        super()
      end

      def chef
        builder(:chef)
        @chef ||= Chef.new
      end

      def shell
        builder(:shell)
        @shell ||= Shell.new
      end

      def git_artifact
        @git_artifact ||= GitArtifact.new
      end

      def docker
        @docker ||= Docker.new
      end

      def name(value)
        @_name = value
      end

      def builder(type)
        type = type.to_sym
        raise "Type `#{type}` isn't supported!" unless [:chef, :shell].include?(type)
        raise "Another builder type '#{_builder}' already used!" if !_builder.nil? and _builder != type
        @_builder = type
      end

      def apps
        @apps.empty? ? [self] : @apps.flatten
      end

      def to_h
        {
          name:         _name,
          builder:      _builder,
          docker:       @docker.to_h,
          git_artifact: @git_artifact.to_h,
          shell:        @shell.to_h,
          chef:         @chef.to_h
        }
      end

      private

      def clone
        self.class.new.tap do |app|
          app.instance_variable_set(:'@_builder', _builder)
          app.instance_variable_set(:'@_home_path', _home_path)
          app.instance_variable_set(:'@docker', docker.clone_with_marshal)             unless @docker.nil?
          app.instance_variable_set(:'@git_artifact', git_artifact.clone_with_marshal) unless @git_artifact.nil?
          app.instance_variable_set(:'@chef', chef.clone_with_marshal)                 unless @chef.nil?
          app.instance_variable_set(:'@shell', shell.clone_with_marshal)               unless @shell.nil?
        end
      end

      def app(sub_name, &blk)
        clone.tap do |app|
          app.instance_variable_set(:'@_name', [_name, sub_name].compact.join('-'))
          app.instance_eval(&blk) if block_given?
          @apps += app.apps
        end
      end
    end
  end
end
