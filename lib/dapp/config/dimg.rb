module Dapp
  module Config
    class Dimg < Base
      attr_reader :_name

      def initialize(name, project:)
        @_name = name
        super(project: project)
      end

      def _name
        File.join(*[_project.name, @_name].compact)
      end

      module InstanceMethods
        attr_accessor :_chef, :_shell, :_docker, :_git_artifact, :_mounts

        def chef(&blk)
          _chef(&blk)
        end

        def shell(&blk)
          _shell(&blk)
        end

        def docker(&blk)
          _docker(&blk)
        end

        def git_artifact(type_or_repo_url, &blk)
          type = (type_or_repo_url.to_sym == :local) ? :local : :remote
          (@_git_artifact ||= GitArtifact.new(project: _project)).send(type, type_or_repo_url, &blk)
        end

        def mount(to, &blk)
          _mounts << Directive.Mount.new(to, project: _project, &blk)
        end

        def _chef(&blk)
          @_chef ||= Directive::Chef.new(project: _project, &blk)
        end

        def _shell(&blk)
          @_shell ||= Directive::Shell.new(project: _project, &blk)
        end

        def _docker(&blk)
          @_docker ||= Directive::Docker.new(project: _project, &blk)
        end

        def _mounts
          @_mounts ||= []
        end

        class GitArtifact < Base
          attr_reader :_local, :_remote

          def initialize(project:)
            @_local = []
            @_remote = []

            super
          end

          def local(_, &blk)
            @_local << Directive::GitArtifactLocal.new(project: _project, &blk)
          end

          def remote(repo_url, &blk)
            @_remote << Directive::GitArtifactRemote.new(repo_url, project: _project, &blk)
          end

          def _local
            @_local.map(&:_exports).flatten
          end

          def _remote
            @_remote.map(&:_exports).flatten
          end
        end
      end
      include InstanceMethods
    end
  end
end
