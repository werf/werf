module Dapp
  module Deployment
    module Config
      module Directive
        class App < Base
          include InstanceMethods
          include Mod::Bootstrap

          attr_reader :_name

          def initialize(name, dapp:)
            self._name = name
            super(dapp: dapp)
          end

          def _name=(name)
            sub_directive_eval do
              return self if name.nil?
              name = name.to_s
              validate_compliance!(hostname_pattern, name, :app_name_incorrect)
              @_name = name
            end
          end
        end
      end
    end
  end
end
