module Dapp
  module Deployment
    module Config
      module Directive
        class Namespace < Base
          module InstanceMethods
            attr_reader :_environment, :_secret_environment, :_scale

            def environment(**kwargs)
              sub_directive_eval { _environment.merge!(**kwargs) }
            end

            def secret_environment(**kwargs)
              sub_directive_eval { _secret_environment.merge!(**kwargs) }
            end

            def scale(value)
              sub_directive_eval do
                value.to_i.tap do |v|
                  raise ::Dapp::Error::Config, code: :unsupported_scale_value, data: { value: value } unless v > 0
                  @_scale = v
                end
              end
            end

            protected

            def namespace_init_variables!
              @_environment        = {}
              @_secret_environment = {}
            end
          end # InstanceMethods
        end # Namespace
      end # Directive
    end # Config
  end # Deployment
end # Dapp
