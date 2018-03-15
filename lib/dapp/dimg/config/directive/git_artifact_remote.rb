module Dapp
  module Dimg
    module Config
      module Directive
        class GitArtifactRemote < GitArtifactLocal
          include ::Dapp::Helper::Url

          attr_reader :_url, :_name, :_branch, :_commit

          def initialize(url, **kwargs, &blk)
            @_url  = url
            @_name = git_url_to_name(url)

            super(**kwargs, &blk)
          end

          def branch(value)
            sub_directive_eval { @_branch = value.to_s }
          end

          def commit(value)
            sub_directive_eval { @_commit = value }
          end

          def _export
            super do |export|
              export._url    = @_url
              export._name   = @_name
              export._branch ||= @_branch
              export._commit ||= @_commit
            end
          end

          class Export < GitArtifactLocal::Export
            attr_accessor :_url, :_name, :_branch, :_commit

            def _artifact_options
              super.merge(name: _name, branch: _branch, commit: _commit)
            end

            def branch(value)
              sub_directive_eval { @_branch = value }
            end

            def commit(value)
              sub_directive_eval { @_commit = value }
            end

            def validate!
              super
              raise Error::Config, code: :git_artifact_remote_branch_with_commit if !_branch.nil? && !_commit.nil?
            end
          end
        end
      end
    end
  end
end
