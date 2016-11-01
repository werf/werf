module Dapp
  module Config
    module Directive
      class Mount < Base
        attr_reader :_from, :_to
        attr_reader :_type

        def initialize(to)
          raise Error::Config, code: :mount_to_absolute_path_required unless Pathname(to).absolute?
          @_to = to.to_s
          super()
        end

        def from(path_or_type)
          path_or_type = path_or_type.to_s
          if [:tmp_dir, :build_dir].include? path_or_type.to_sym
            @_type = path_or_type.to_sym
          else
            raise Error::Config, code: :mount_from_absolute_path_required unless Pathname(path_or_type).absolute?
            @_from = path_or_type
          end
        end
      end
    end
  end
end
