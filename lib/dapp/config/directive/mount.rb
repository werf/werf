module Dapp
  module Config
    module Directive
      class Mount < Base
        attr_reader :_from, :_to
        attr_reader :_type

        def initialize(to, project:)
          @_to = to.to_s
          super(project: project)
        end

        def from(path_or_type)
          path_or_type = path_or_type.to_s
          if [:tmp_dir, :build_dir].include? path_or_type.to_sym
            @_type = path_or_type.to_sym
          else
            @_type = :base
            raise if Pathname(path_or_type).absolute? # TODO: absolute required
            @_from = path_or_type
          end
        end
      end
    end
  end
end
