module Dapp
  module Dimg
    module Config
      module Directive
        class Mount < Base
          attr_reader :_to
          attr_reader :_from
          attr_reader :_type

          def initialize(to, **kwargs, &blk)
            raise Error::Config, code: :mount_to_absolute_path_required unless Pathname((to = to.to_s)).absolute?
            @_to = path_format(to)

            super(**kwargs, &blk)
          end

          def from(type)
            sub_directive_eval do
              type = type.to_sym
              raise Error::Config, code: :mount_from_type_required unless [:tmp_dir, :build_dir].include? type
              @_type = type
            end
          end

          def from_path(path)
            sub_directive_eval do
              @_from = path_format(path)
              @_type = :custom_dir
            end
          end

          def validate!
            raise Error::Config, code: :mount_from_or_from_path_required if _type.nil?
          end
        end
      end
    end
  end
end
