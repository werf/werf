module Dapp
  module Dimg
    module Dapp
      module Dimg
        def dimg(config:, **kwargs)
          dimg_after_define_hook(config: config, **kwargs) do
            (@dimg ||= {})[config._name] ||= ::Dapp::Dimg::Dimg.new(config: config, dapp: self, **kwargs)
          end
        end

        def artifact_dimg(config:, **kwargs)
          dimg_after_define_hook(config: config, **kwargs) do
            (@artifacts_dimgs ||= {})[config._name] ||= ::Dapp::Dimg::Artifact.new(config: config, dapp: self, **kwargs)
          end
        end

        def dimg_after_define_hook(**kwargs)
          should_be_built = kwargs[:should_be_built] || false
          yield.tap do |dimg|
            if should_be_built && dimg.should_be_built != should_be_built
              dimg.enable_should_be_built
              dimg.should_be_built!
            end
          end
        end

        def dimg_layer(config:, **kwargs)
          (@dimg_layers ||= {})[config._name] ||= ::Dapp::Dimg::Dimg.new(config: config, dapp: self, **kwargs)
        end

        def artifact_dimg_layer(config:, **kwargs)
          (@artifact_dimg_layers ||= {})[config._name] ||= ::Dapp::Dimg::Artifact.new(config: config, dapp: self, **kwargs)
        end

        def _terminate_dimg_on_terminate(dimg)
          @_call_before_terminate << proc{dimg.terminate}
        end
      end # Dimg
    end # Dapp
  end # Dimg
end # Dapp
