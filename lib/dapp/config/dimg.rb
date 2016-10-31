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
        attr_accessor :_chef, :_shell, :_docker, :_git_artifacts, :_mounts

        def chef(&blk)
          _chef(&blk)
        end

        def shell(&blk)
          _shell(&blk)
        end

        def docker(&blk)
          _docker(&blk)
        end

        def git_artifact(type_or_git_repo, &blk)
          type = (type_or_git_repo == :local) ? type_or_git_repo : :remote
          _git_artifacts[type] << begin
            if type == :local
              Directive::GitArtifactLocal.new(project: _project, &blk)
            elsif type == :remote
              Directive::GitArtifactRemote.new(type_or_git_repo, project: _project, &blk)
            end
          end
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

        def _git_artifacts
          @_git_artifacts ||= { local: [], remote: [] }
        end

        def _mounts
          @_mounts ||= []
        end
      end
      include InstanceMethods
    end
  end
end
