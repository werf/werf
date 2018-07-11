module Dapp
  module Dimg
    module Config
      module Directive
        class GitArtifactRemote < GitArtifactLocal
          include ::Dapp::Helper::Url

          attr_reader :_url, :_name, :_branch, :_tag, :_commit

          def initialize(url, **kwargs, &blk)
            @_url  = url
            @_name = git_url_to_name(url)

            super(**kwargs, &blk)
          end

          def branch(value)
            sub_directive_eval { @_branch = value.to_s }
          end

          def tag(value)
            sub_directive_eval { @_tag = value.to_s }
          end

          def commit(value)
            sub_directive_eval { @_commit = value.to_s }
          end

          def _export
            super do |export|
              export._url    = @_url
              export._name   = @_name
              export._branch ||= @_branch
              export._tag    ||= @_tag
              export._commit ||= @_commit

              yield(export) if block_given?
            end
          end

          class Export < GitArtifactLocal::Export
            attr_accessor :_url, :_name, :_branch, :_tag, :_commit

            def _artifact_options
              super.merge(name: _name, branch: _branch, tag: _tag, commit: _commit)
            end

            def branch(value)
              sub_directive_eval { @_branch = value.to_s }
            end

            def tag(value)
              sub_directive_eval { @_tag = value.to_s }
            end

            def commit(value)
              sub_directive_eval { @_commit = value.to_s }
            end

            def validate!
              super
              refs = [_branch, _tag, _commit].compact
              raise ::Dapp::Error::Config, code: :git_artifact_remote_with_refs if refs.length > 1
            end
          end
        end
      end
    end
  end
end
