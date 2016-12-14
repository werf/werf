module Dapp
  module Config
    class Dimg < Base
      # InstanceMethods
      module InstanceMethods
        attr_reader :_builder
        attr_reader :_chef, :_shell, :_docker, :_git_artifact, :_mount, :_artifact
        attr_reader :_install_dependencies, :_setup_dependencies

        def dev_mode
          @_dev_mode = true
        end

        def install_depends_on(*args)
          _install_dependencies.concat(args)
        end

        def setup_depends_on(*args)
          _setup_dependencies.concat(args)
        end

        def chef(&blk)
          builder(:chef)
          directive_eval(_chef, &blk)
        end

        def shell(&blk)
          builder(:shell)
          directive_eval(_shell, &blk)
        end

        def docker(&blk)
          directive_eval(_docker, &blk)
        end

        def artifact(&blk)
          _artifact.concat begin
                             pass_to(ArtifactGroup.new(project: project)).tap do |artifact_group|
                               artifact_group.instance_eval(&blk) if block_given?
                             end._export
                           end
        end

        def git(url = nil, &blk)
          type = url.nil? ? :local : :remote
          _git_artifact.send(type, url.to_s, &blk)
        end

        def mount(to, &blk)
          _mount << Directive::Mount.new(to, project: project, &blk)
        end

        def _dev_mode
          !!@_dev_mode
        end

        def _builder
          @_builder || :none
        end

        def _chef
          @_chef ||= Directive::Chef.new(project: project)
        end

        def _shell
          @_shell ||= Directive::Shell::Dimg.new(project: project)
        end

        def _docker
          @_docker ||= Directive::Docker::Dimg.new(project: project)
        end

        def _mount
          @_mount ||= []
        end

        def _git_artifact
          @_git_artifact ||= GitArtifact.new(project: project)
        end

        [:build_dir, :tmp_dir].each do |mount_type|
          define_method "_#{mount_type}_mount" do
            _mount.select { |m| m._type == mount_type }
          end
        end

        def _install_dependencies
          @_install_dependencies ||= []
        end

        def _setup_dependencies
          @_setup_dependencies ||= []
        end

        def _artifact
          @_artifact ||= []
        end

        [:before, :after].each do |order|
          [:setup, :install].each do |stage|
            define_method "_#{order}_#{stage}_artifact" do
              _artifact.select do |art|
                art.public_send("_#{order}") == stage
              end
            end
          end
        end

        def _import_artifact
          _artifact.select(&:not_associated?)
        end

        # GitArtifact
        class GitArtifact < Directive::Base
          attr_reader :_local, :_remote

          def initialize(**kwargs, &blk)
            @_local = []
            @_remote = []

            super(**kwargs, &blk)
          end

          def local(_, &blk)
            @_local << Directive::GitArtifactLocal.new(project: project, &blk)
          end

          def remote(repo_url, &blk)
            @_remote << Directive::GitArtifactRemote.new(repo_url, project: project, &blk)
          end

          def _local
            @_local.map(&:_export).flatten
          end

          def _remote
            @_remote.map(&:_export).flatten
          end

          protected

          def empty?
            (_local + _remote).empty?
          end

          def validate!
            (_local + _remote).each { |exp| exp.send(:validate!) }
          end
        end

        protected

        def builder(type)
          @_builder = type if _builder == :none
          raise Error::Config, code: :builder_type_conflict unless @_builder == type
        end

        def directive_eval(directive, &blk)
          directive.instance_eval(&blk) if block_given?
          directive
        end

        def pass_to(obj)
          passed_directives.each do |directive|
            directive_variable_name = :"@_#{directive}"

            next if (value = instance_variable_get(directive_variable_name)).nil?
            obj_value = obj.instance_variable_get(directive_variable_name)

            if value.is_a?(Directive::Base)
              obj.builder(directive) if [:chef, :shell].include? directive
              if obj_value.nil?
                obj.instance_variable_set(directive_variable_name, value.send(:_clone))
              else
                obj_value.send(:merge, value)
              end
            elsif respond_to?(:"merge_#{directive}", true)
              obj.instance_variable_set(directive_variable_name, send(:"merge_#{directive}", obj_value, value))
            else
              raise
            end
          end
          obj
        end

        def passed_directives
          [:chef, :shell, :docker,
           :git_artifact, :mount,
           :artifact, :builder, :dev_mode,
           :install_dependencies, :setup_dependencies]
        end
      end
    end
  end
end
