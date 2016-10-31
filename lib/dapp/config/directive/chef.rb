module Dapp
  module Config
    module Directive
      class Chef < Base
        attr_accessor :_dimod, :_recipe, :_attributes

        def initialize
          @_dimod = []
          @_recipe = []

          super
        end

        def dimod(*args)
          @_dimod.concat(args)
        end

        def recipe(*args)
          @_recipe.concat(args)
        end

        def attributes
          @_attributes ||= Attributes.new
        end

        %i(before_install install before_setup setup build_artifact).each do |stage|
          define_method("#{stage}_attributes") do
            var = "@#{stage}_attributes"
            instance_variable_get(var) || instance_variable_set(var, Attributes.new)
          end
        end

        protected

        %i(before_install install before_setup setup build_artifact).each do |stage|
          define_method("_#{stage}_attributes") do
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
