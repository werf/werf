module Dapp
  module Config
    module Directive
      # Mount
      class Mount < Base
        attr_reader :_from, :_to
        attr_reader :_type

        def initialize(to, **kwargs, &blk)
          raise Error::Config, code: :mount_to_absolute_path_required unless Pathname((to = to.to_s)).absolute?
          @_to = to

          super(**kwargs, &blk)
        end

        protected

        def from(path_or_type)
          path_or_type = path_or_type.to_sym
          raise Error::Config, code: :mount_from_type_required unless [:tmp_dir, :build_dir].include? path_or_type
          @_type = path_or_type
        end
      end
    end
  end
end
