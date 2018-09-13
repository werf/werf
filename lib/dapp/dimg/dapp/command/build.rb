module Dapp
  module Dimg
    module Dapp
      module Command
        module Build
          def build
            build_configs.each do |config|
              log_dimg_name_with_indent(config) do
                dimg(config: config).build!
              end
            end
          rescue ::Dapp::Error::Shellout, Error::Default
            build_context_export unless options[:build_context_directory].nil?
            raise
          end
        end
      end
    end
  end # Dimg
end # Dapp
