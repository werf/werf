module Dapp
  module Config
    module Directive
      # GitArtifactRemote
      class GitArtifactRemote < GitArtifactLocal
        attr_reader :_url, :_name, :_branch, :_commit

        def initialize(url, **kwargs, &blk)
          @_url  = url
          @_name = url.gsub(%r{.*?([^\/ ]+)\.git}, '\\1')

          super(**kwargs, &blk)
        end

        def _export
          super do |export|
            export._url    = @_url
            export._name   = @_name
            export._branch ||= @_branch
            export._commit ||= @_commit
          end
        end

        # Export
        class Export < GitArtifactLocal::Export
          attr_accessor :_url, :_name, :_branch, :_commit

          def _artifact_options
            super.merge(name: _name, branch: _branch, commit: _commit)
          end

          protected

          def branch(value)
            @_branch = value
          end

          def commit(value)
            @_commit = value
          end

          def validate!
            super
            raise Error::Config, code: :git_artifact_remote_branch_with_commit if !_branch.nil? && !_commit.nil?
          end
        end

        protected

        def branch(value)
          @_branch = value
        end

        def commit(value)
          @_commit = value
        end
      end
    end
  end
end
