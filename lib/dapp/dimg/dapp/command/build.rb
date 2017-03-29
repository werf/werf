module Dapp
  module Dimg
    module Dapp
      module Command
        module Build
          def build
            setup_ssh_agent

            build_configs.each do |config|
              log_dimg_name_with_indent(config) do
                Dimg.new(config: config, dapp: self).build!
              end
            end
          rescue ::Dapp::Error::Shellout, Error::Base
            build_context_export unless cli_options[:build_context_directory].nil?
            raise
          end
        end
      end
    end
  end # Dimg
end # Dapp
