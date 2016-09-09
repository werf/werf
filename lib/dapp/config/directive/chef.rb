module Dapp
  module Config
    module Directive
      # Chef
      class Chef
        attr_reader :_modules
        attr_reader :_recipes
        attr_reader :_attributes

        def initialize
          @_modules = []
          @_recipes = []
          @_attributes = []
        end

        def module(*args)
          @_modules.concat(args)
        end

        def reset_modules
          @_modules.clear
        end

        def skip_module(*args)
          @_modules.reject! { |mod| args.include? mod }
        end

        def recipe(*args)
          @_recipes.concat(args)
        end

        def remove_recipe(*args)
          @_recipes.reject! { |recipe| args.include? recipe }
        end

        def reset_recipes
          @_recipes.clear
        end

        def attribute(*args)
          @_attributes.concat(args)
        end

        def remove_attribute(*args)
          @_attributes.reject! { |attribute| args.include? attribute }
        end

        def reset_attributes
          @_attributes.clear
        end

        def reset_all
          reset_modules
          reset_recipes
          reset_attributes
        end

        protected

        def clone
          Marshal.load(Marshal.dump(self))
        end

        def empty?
          @_modules.empty? && @_recipes.empty?
        end
      end
    end
  end
end
