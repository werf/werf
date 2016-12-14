module Dapp
  module Config
    module Directive
      # Base
      class Base < Config::Base
        protected

        def clone
          _clone
        end

        def merge(obj)
          cloned_obj = obj.clone

          instance_variables.each do |variable_name|
            next if (obj_value = cloned_obj.instance_variable_get(variable_name)).nil?
            value = instance_variable_get(variable_name)

            case obj_value
            when Directive::Base
              if value.nil?
                instance_variable_set(variable_name, obj_value)
              else
                value.send(:merge, obj_value)
              end
            when Array then instance_variable_set(variable_name, obj_value.concat(Array(value)))
            when Hash then instance_variable_set(variable_name, obj_value.merge(Hash(value)))
            else
              instance_variable_set(variable_name, value || obj_value)
            end
          end
        end
      end
    end
  end
end
