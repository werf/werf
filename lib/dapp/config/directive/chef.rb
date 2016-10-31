module Dapp
  module Config
    module Directive
      class Chef < Base
        attr_accessor :_module, :_recipe

        def initialize
          @_module = []
          @_recipe = []

          super
        end

        def module(*args)
          @_module.concat(args)
        end

        def recipe(*args)
          @_recipe.concat(args)
        end

        def attributes
          @attributes ||= Attributes.new
        end

        %i(before_install install before_setup setup build_artifact).each do |stage|
          define_method("_#{stage}_attributes") do
            var = "@#{stage}_attributes"
            instance_variable_get(var) || instance_variable_set(var, Attributes.new)
          end
        end

        protected

        %i(before_install install before_setup setup build_artifact).each do |stage|
          define_method("#{stage}_attributes") do
            attributes.in_depth_merge send("#{stage}_attributes")
          end
        end

        # Attributes
        class Attributes < Hash
          def [](key)
            super || begin
              self[key] = self.class.new
            end
          end
        end # Attributes
      end
    end
  end
end
