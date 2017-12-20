module Dapp
  module Dimg
    module Config
      module Directive
        class ArtifactBase < Base
          attr_reader :_owner, :_group

          def initialize(**kwargs, &blk)
            @_export = []
            super(**kwargs, &blk)
          end

          def owner(owner)
            sub_directive_eval { @_owner = owner }
          end

          def group(group)
            sub_directive_eval { @_group = group }
          end

          def export(absolute_dir_path = nil, &blk)
            self.class.const_get('Export').new(absolute_dir_path, dapp: dapp, &blk).tap do |export|
              @_export << export
            end
          end

          def _export
            @_export.each do |export|
              export._owner ||= @_owner
              export._group ||= @_group

              yield(export) if block_given?
            end
          end

          class Export < Directive::Base
            attr_accessor :_cwd, :_to, :_include_paths, :_exclude_paths, :_owner, :_group

            def initialize(cwd, **kwargs, &blk)
              self._cwd = cwd
              @_include_paths ||= []
              @_exclude_paths ||= []

              super(**kwargs, &blk)
            end

            def _artifact_options
              {
                to:            _to,
                cwd:           _cwd,
                include_paths: _include_paths,
                exclude_paths: _exclude_paths,
                owner:         _owner,
                group:         _group
              }
            end

            def _cwd
              @_cwd ||= _to
            end

            def _cwd=(value)
              return if value.nil?
              raise ::Dapp::Error::Config, code: :export_cwd_absolute_path_required unless Pathname(value).absolute?
              @_cwd = path_format(value)
            end

            def to(absolute_path)
              sub_directive_eval do
                raise ::Dapp::Error::Config, code: :export_to_absolute_path_required unless Pathname(absolute_path).absolute?
                @_to = path_format(absolute_path)
              end
            end

            def include_paths(*relative_paths)
              sub_directive_eval do
                unless relative_paths.all? { |path| Pathname(path).relative? }
                  raise ::Dapp::Error::Config, code: :export_include_paths_relative_path_required
                end
                _include_paths.concat(relative_paths.map(&method(:path_format)))
              end
            end

            def exclude_paths(*relative_paths)
              sub_directive_eval do
                unless relative_paths.all? { |path| Pathname(path).relative? }
                  raise ::Dapp::Error::Config, code: :export_exclude_paths_relative_path_required
                end
                _exclude_paths.concat(relative_paths.map(&method(:path_format)))
              end
            end

            def owner(owner)
              sub_directive_eval { @_owner = owner }
            end

            def group(group)
              sub_directive_eval { @_group = group }
            end

            def validate!
              raise ::Dapp::Error::Config, code: :export_to_required if _to.nil?
            end
          end
        end
      end
    end
  end
end
