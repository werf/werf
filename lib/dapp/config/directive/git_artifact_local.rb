module Dapp
  module Config
    module Directive
      class GitArtifactLocal < Base
        attr_reader :_owner, :_group
        attr_reader :_exports

        def initialize(project:)
          @_exports = []
          super
        end

        def owner(owner)
          @_owner = owner
        end

        def group(group)
          @_group = group
        end

        def export(absolute_dir_path = '/', &blk)
          @_exports << self.class.const_get('Export').new(absolute_dir_path, project: _project, &blk)
        end

        def _exports
          @_exports.each do |export|
            export._owner ||= @_owner
            export._group ||= @_group

            yield(export) if block_given?
          end
        end

        protected

        class Export < Directive::Base
          attr_accessor :_cwd, :_include_paths, :_exclude_paths, :_owner, :_group

          def initialize(cwd, project:)
            raise unless Pathname(cwd).absolute? # TODO: absolute required
            @_cwd = cwd
            @_include_paths ||= []
            @_exclude_paths ||= []

            super(project: project)
          end

          def to(absolute_path)
            raise unless Pathname(absolute_path).absolute? # TODO: absolute required
            @_to = absolute_path
          end

          def include_paths(*relative_paths)
            raise unless relative_paths.all? { |path| Pathname(path).relative? } # TODO: relative required
            _include_paths.concat(relative_paths)
          end

          def exclude_paths(*relative_paths)
            raise unless relative_paths.all? { |path| Pathname(path).relative? } # TODO: relative required
            _exclude_paths.concat(relative_paths)
          end

          def owner(owner)
            @_owner = owner
          end

          def group(group)
            @_group = group
          end
        end
      end
    end
  end
end
