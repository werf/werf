module Dapp
  module Dimg
    module Config
      module Directive
        module Docker
          class Artifact < Base
            def method_missing(m, *args)
              raise Error::Config, code: :docker_artifact_unsupported_directive if Dimg.instance_methods.include?(m)
              super
            end

            def _change_options
              {}
            end
          end
        end
      end
    end
  end
end
