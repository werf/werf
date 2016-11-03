module Dapp
  module Config
    class Dimg < Base
      module InstanceMethods
        attr_reader :_builder
        attr_reader :_chef, :_shell, :_docker, :_git_artifact, :_mount, :_artifact
        attr_reader :_install_dependencies, :_setup_dependencies

        def install_depends_on(*args)
          @_install_dependencies.concat(args)
        end

        def setup_depends_on(*args)
          @_setup_dependencies.concat(args)
        end

        def chef(&blk)
          builder_validation(:chef)
          directive_eval(_chef, &blk)
        end

        def shell(&blk)
          builder_validation(:shell)
          directive_eval(_shell, &blk)
        end

        def docker(&blk)
          directive_eval(_docker, &blk)
        end

        def artifact(&blk)
          _artifact.concat begin
                             pass_to_custom(ArtifactGroup.new(project: project), :clone_to_artifact).tap do |artifact_group|
                               artifact_group.instance_eval(&blk) if block_given?
                             end._export
                           end
        end

        def git_artifact(type_or_repo_url, &blk)
          type = (type_or_repo_url.to_sym == :local) ? :local : :remote
          _git_artifact.send(type, type_or_repo_url, &blk)
        end

        def mount(to, &blk)
          _mount << Directive::Mount.new(to, &blk)
        end

        def _builder
          @_builder || :none
        end

        def _chef
          @_chef ||= Directive::Chef.new
        end

        def _shell
          @_shell ||= Directive::Shell::Dimg.new
        end

        def _docker
          @_docker ||= Directive::Docker::Dimg.new
        end

        def _mount
          @_mount ||= []
        end

        def _git_artifact
          @_git_artifact ||= GitArtifact.new
        end

        [:build_dir, :tmp_dir].each do |mount_type|
          define_method "_#{mount_type}_mount" do
            _mount.select { |m| m._type == mount_type }
          end
        end

        def _custom_mount
          _mount.select { |m| m._type.nil? }
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

        class GitArtifact
          attr_reader :_local, :_remote

          def initialize
            @_local = []
            @_remote = []
          end

          def local(_, &blk)
            @_local << Directive::GitArtifactLocal.new(&blk)
          end

          def remote(repo_url, &blk)
            @_remote << Directive::GitArtifactRemote.new(repo_url, &blk)
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

        def builder_validation(type)
          @_builder = type if _builder == :none
          raise Error::Config, code: :builder_type_conflict unless @_builder == type
        end

        def directive_eval(directive, &blk)
          directive.instance_eval(&blk) if block_given?
        end

        def pass_to_default(dimg)
          pass_to_custom(dimg, :clone)
        end

        def pass_to_custom(obj, clone_method)
          passed_directives.each do |directive|
            next if (variable = instance_variable_get(directive)).nil?

            obj.instance_variable_set(directive, begin
              case variable
              when Directive::Base, GitArtifact then variable.send(clone_method)
              when Symbol then variable
              when Array then variable.dup
              else
                raise
              end
            end)
          end
          obj
        end

        def passed_directives
          [:@_chef, :@_shell, :@_docker, :@_git_artifact, :@_mount, :@_artifact, :@_builder]
        end
      end
    end
  end
end
