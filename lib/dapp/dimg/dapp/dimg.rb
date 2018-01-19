module Dapp
  module Dimg
    module Dapp
      module Dimg
        def dimg(config:, **kwargs)
          (@dimgs ||= {})[config._name] ||= ::Dapp::Dimg::Dimg.new(config: config, dapp: self, **kwargs)
        end

        def artifact_dimg(config:, **kwargs)
          (@artifacts_dimgs ||= {})[config._name] ||= ::Dapp::Dimg::Artifact.new(config: config, dapp: self, **kwargs)
        end
      end # Dimg
    end # Dapp
  end # Dimg
end # Dapp
