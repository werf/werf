module Dapp
  module Config
    class Dimg < Base
      attr_reader :_name

      def initialize(name, project:)
        @_name = name
        super(project: project)
      end

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
          _chef(&blk)
        end

        def shell(&blk)
          builder_validation(:shell)
          _shell(&blk)
        end

        def docker(&blk)
          _docker(&blk)
        end

        def artifact(&blk)
          _artifact.concat begin
                             pass_to_custom(ArtifactGroup.new(project: project), :clone_to_artifact).tap do |artifact_group|
                               artifact_group.instance_eval(&blk) if block_given?
                             end._export
                           end
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

        def git_artifact(type_or_repo_url, &blk)
          type = (type_or_repo_url.to_sym == :local) ? :local : :remote
          (@_git_artifact ||= GitArtifact.new).send(type, type_or_repo_url, &blk)
        end

        def mount(to, &blk)
          _mount << Directive::Mount.new(to, &blk)
        end

        def _chef(&blk)
          @_chef ||= Directive::Chef.new(&blk)
        end

        def _shell(&blk)
          @_shell ||= Directive::Shell::Dimg.new(&blk)
        end

        def _docker(&blk)
          @_docker ||= Directive::Docker::Dimg.new(&blk)
        end

        def _mount
          @_mount ||= []
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
        end

        protected

        def builder_validation(type)
          @_builder ||= type
          raise Error::Config, code: :builder_type_conflict unless _builder == type
        end
      end
      include InstanceMethods
    end
  end
end
