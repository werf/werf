module Dapp
  module Dimg
    class Dapp
      module Command
        module Build
          def build
            setup_ssh_agent

            build_configs.each do |config|
              log_dimg_name_with_indent(config) do
                Dimg.new(config: config, dapp: self).build!
              end
            end
          end
        end
      end
    end
  end # Dimg
end # Dapp
