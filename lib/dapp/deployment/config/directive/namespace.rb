module Dapp
  module Deployment
    module Config
      module Directive
        class Namespace < Base
          include InstanceMethods

          attr_reader :_name

          def initialize(name, dapp:)
            self._name = name
            super(dapp: dapp)
          end

          def _name=(name)
            sub_directive_eval do
              name = name.to_s
              validate_compliance!(hostname_pattern, name, :namespace_name_incorrect)
              @_name = name
            end
          end
        end
      end
    end
  end
end
