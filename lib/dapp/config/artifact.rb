module Dapp
  module Config
    # Artifact
    module Artifact
      # Base
      class Base
        attr_accessor :_where_to_add, :_cwd, :_paths, :_owner, :_group

        def initialize(where_to_add, **options)
          @_cwd          = ''
          @_where_to_add = where_to_add

          options.each do |k, v|
            respond_to?("_#{k}=") ? send(:"_#{k}=", v) : raise(Error::Config, code: code,
                                                                              data: { type: object_name, attr: k })
          end
        end

        def _paths
          Array(@_paths)
        end

        def _artifact_options
          {
            where_to_add: _where_to_add,
            cwd:          _cwd,
            paths:        _paths,
            owner:        _owner,
            group:        _group
          }
        end

        def clone
          Marshal.load(Marshal.dump(self))
        end

        protected

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
        attr_accessor :_before, :_after

        def _before=(stage)
          @_before = stage.to_sym
        end

        def _after=(stage)
          @_after = stage.to_sym
        end

        protected

        def code
          :artifact_unexpected_attribute
        end
      end
    end
  end
end
