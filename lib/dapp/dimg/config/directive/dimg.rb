module Dapp
  module Dimg
    module Config
      module Directive
        class Dimg < Base
          include Validation
          include InstanceMethods

          attr_reader :_name

          def initialize(name, dapp:)
            self._name = name
            super(dapp: dapp)
          end

          def _name=(name)
            sub_directive_eval do
              return self if name.nil?
              name = name.to_s
              validate_compliance!(dimg_name_pattern, name, :dimg_name_incorrect)
              @_name = name
            end
          end

          protected

          def dimg_name_pattern
            separator = '[_\.]|(__)|(-*)'
            alpha_numeric = '[a-z0-9]'
            "^#{alpha_numeric}(#{separator}#{alpha_numeric})*$"
          end
        end
      end
    end
  end
end
