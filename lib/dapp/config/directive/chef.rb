module Dapp
  module Config
    module Directive
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

        def attributes
          @attributes ||= new_attributes
        end

        %i(before_install install before_setup setup build_artifact).each do |stage|
          define_method("#{stage}_attributes") do
            "@#{stage}_attributes".tap do |var|
              instance_variable_get(var) || instance_variable_set(var, new_attributes)
            end
          end

          define_method("_#{stage}_attributes") do
            attributes.in_depth_merge send("#{stage}_attributes")
          end

          define_method("reset_#{stage}_attributes") do
            instance_variable_set("@#{stage}_attributes", nil)
          end
        end

        def reset_attributes
          @attributes = nil
        end

        def reset_all_attributes
          reset_attributes
          %i(before_install install before_setup setup build_artifact).each do |stage|
            send("reset_#{stage}_attributes")
          end
        end

        def reset_all
          reset_modules
          reset_recipes
          reset_all_attributes
        end

        protected

        def clone
          Marshal.load(Marshal.dump(self))
        end

        def empty?
          @_modules.empty? && @_recipes.empty?
        end

        def new_attributes
          Hash.new { |hash, key| hash[key] = Hash.new &hash.default_proc }
        end
      end
    end
  end
end
