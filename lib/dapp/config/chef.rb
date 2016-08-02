module Dapp
  module Config
    # Chef
    class Chef
      attr_reader :_modules
      attr_reader :_recipes

      def initialize
        @_modules = []
        @_recipes = []
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

      def reset_all
        reset_modules
        reset_recipes
      end

      def clone
        Marshal.load(Marshal.dump(self))
      end
    end
  end
end
