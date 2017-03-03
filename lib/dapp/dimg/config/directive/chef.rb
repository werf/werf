module Dapp
  module Dimg
    module Config
      module Directive
        class Chef < Base
          attr_accessor :_dimod, :_cookbook, :_recipe, :_attributes

          def initialize(**kwargs, &blk)
            @_dimod = []
            @_recipe = []
            @_cookbook = {}

            super(**kwargs, &blk)
          end

          def dimod(name, *args)
            @_dimod << name
            cookbook(name, *args)
          end

          def recipe(name)
            @_recipe << name
          end

          def attributes
            @_attributes ||= Attributes.new
          end

          def cookbook(name, version_constraint=nil, **kwargs)
            @_cookbook[name] = {}.tap do |desc|
              desc.update(**kwargs)
              desc[:name] = name
              desc[:version_constraint] = version_constraint if version_constraint
            end
          end

          %i(before_install install before_setup setup build_artifact).each do |stage|
            define_method("_#{stage}_attributes") do
              var = "@__#{stage}_attributes"
              instance_variable_get(var) || instance_variable_set(var, Attributes.new)
            end
          end

          class Attributes < Hash
            def [](key)
              super || begin
                self[key] = self.class.new
              end
            end
          end # Attributes

          %i(before_install install before_setup setup build_artifact).each do |stage|
            define_method("__#{stage}_attributes") do
              attributes.in_depth_merge send("_#{stage}_attributes")
            end
          end

          protected

          def empty?
            (@_dimod + @_recipe).empty? && attributes.empty?
          end
        end
      end
    end
  end
end
