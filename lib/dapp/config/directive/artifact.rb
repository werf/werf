module Dapp
  module Config
    module Directive
      # Artifact
      module Artifact
        # Base
        class Base
          attr_accessor :_where_to_add, :_cwd, :_paths, :_exclude_paths, :_owner, :_group

          def initialize(where_to_add, **options)
            @_cwd          = ''
            @_where_to_add = where_to_add

            options.each do |k, v|
              respond_to?("_#{k}=") ? send(:"_#{k}=", v) : raise(Error::Config, code: code,
                                                                 data: { type: object_name, attr: k })
            end
          end

          def _paths
            base_paths(@_paths)
          end

          def _exclude_paths
            base_paths(@_exclude_paths)
          end

          def _artifact_options
            {
              where_to_add:  _where_to_add,
              cwd:           _cwd,
              paths:         _paths,
              exclude_paths: _exclude_paths,
              owner:         _owner,
              group:         _group
            }
          end

          protected

          def clone
            Marshal.load(Marshal.dump(self))
          end

          def base_paths(paths)
            Array(paths)
          end

          def code
            raise
          end

          def object_name
            self.class.to_s.split('::').last
          end
        end

        # Stage
        class Stage < Base
          attr_accessor :_config

          protected

          def clone
            artifact_options = Marshal.load(Marshal.dump(_artifact_options))
            where_to_add = artifact_options.delete(:where_to_add)
            self.class.new(where_to_add, config: _config, **artifact_options)
          end

          def code
            :artifact_unexpected_attribute
          end
        end
      end
    end
  end
end
