module Dapp
  module Config
    module Directive
      class Base
        def initialize(dapp:, &blk)
          @dapp = dapp
          initialize_variables
          instance_eval(&blk) if block_given?
        end

        def clone
          _clone
        end

        protected

        attr_reader :dapp

        def path_format(path)
          path = path.to_s
          path = path.chomp('/') unless path == '/'
          path
        end

        def validate_compliance!(pattern, value, error_code)
          raise Error::Config, code: error_code, data: { value: value, pattern: pattern } unless /^#{pattern}$/ =~ value
        end

        def initialize_variables
          do_all!('_init_variables!')
        end

        def do_all!(postfix)
          methods
            .select { |m| m.to_s.end_with? postfix }
            .each(&method(:send))
        end

        def directive_eval(directive, &blk)
          directive.instance_eval(&blk) if block_given?
          directive
        end

        def sub_directive_eval
          yield if block_given?
          self
        end

        def pass_to(obj, clone_method = :clone)
          passed_directives.each do |directive|
            next if (variable = instance_variable_get(directive)).nil?

            obj.instance_variable_set(directive, begin
              case variable
              when Base then variable.public_send(clone_method)
              when String, Symbol, Integer, TrueClass, FalseClass then variable
              when Array, Hash then marshal_clone(variable)
              else
                raise
              end
            end)
          end
          obj
        end

        def passed_directives
          []
        end

        def ref_variables
          [:@dapp]
        end

        def marshal_dump
          instance_variables
            .reject {|variable| ref_variables.include? variable}
            .map {|variable| [variable, instance_variable_get(variable)]}
        end

        def marshal_load(variable_values)
          variable_values.each do |variable, value|
            instance_variable_set(variable, value)
          end

          self
        end

        def _clone
          marshal_clone(self).tap do |obj|
            _set_ref_variables_to(obj)
          end
        end

        def _clone_to(obj)
          obj.tap do
            obj.marshal_load(marshal_dump)
            _set_ref_variables_to(obj)
          end
        end

        def _set_ref_variables_to(obj)
          ref_variables.each do |ref_variable|
            obj.instance_variable_set(ref_variable, instance_variable_get(ref_variable))
          end
        end

        def marshal_clone(obj)
          Marshal.load(Marshal.dump(obj))
        end
      end
    end
  end
end
