module Dapp
  module Config
    module Directive
      # ArtifactBase
      class ArtifactBase < Base
        attr_reader :_owner, :_group

        def initialize(**kwargs, &blk)
          @_export = []

          super(**kwargs, &blk)
        end

        def _export
          @_export.each do |export|
            export._owner ||= @_owner
            export._group ||= @_group

            yield(export) if block_given?
          end
        end

        # Export
        class Export < Directive::Base
          attr_accessor :_cwd, :_to, :_include_paths, :_exclude_paths, :_owner, :_group

          def initialize(cwd = '/', **kwargs, &blk)
            raise Error::Config, code: :export_cwd_absolute_path_required unless Pathname(cwd).absolute?
            @_cwd = path_format(cwd)
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

          protected

          def to(absolute_path)
            raise Error::Config, code: :export_to_absolute_path_required unless Pathname(absolute_path).absolute?
            @_to = path_format(absolute_path)
          end

          def include_paths(*relative_paths)
            raise Error::Config, code: :export_include_paths_relative_path_required unless relative_paths.all? { |path| Pathname(path).relative? }
            _include_paths.concat(relative_paths.map(&method(:path_format)))
          end

          def exclude_paths(*relative_paths)
            raise Error::Config, code: :export_exclude_paths_relative_path_required unless relative_paths.all? { |path| Pathname(path).relative? }
            _exclude_paths.concat(relative_paths.map(&method(:path_format)))
          end

          def owner(owner)
            @_owner = owner
          end

          def group(group)
            @_group = group
          end

          def validate!
            raise Error::Config, code: :export_to_required if _to.nil?
          end
        end

        protected

        def owner(owner)
          @_owner = owner
        end

        def group(group)
          @_group = group
        end

        def export(absolute_dir_path = '/', &blk)
          @_export << self.class.const_get('Export').new(absolute_dir_path, project: project, &blk)
        end
      end
    end
  end
end
