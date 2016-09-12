module Dapp
  # Project
  class Project
    # Deps
    module Deps
      # Base
      module Base
        BASE_IMAGE = 'dappdeps/base:0.1.0'.freeze

        def base_container_name # FIXME: hashsum(image) or dockersafe()
          BASE_IMAGE.tr('/', '_').tr(':', '_')
        end

        def base_container
          @base_container ||= begin
            if shellout("docker inspect #{base_container_name}").exitstatus.nonzero?
              log_secondary_process(t(code: 'process.base_container_loading'), short: true) do
                shellout ['docker run',
                          '--restart=no',
                          "--name #{base_container_name}",
                          "--volume /.dapp/deps/base #{BASE_IMAGE}",
                          '2>/dev/null'].join(' ')
              end
            end
            base_container_name
          end
        end

        def bash_path
          '/.dapp/deps/base/bin/bash'
        end
      end # Base
    end # Deps
  end # Project
end # Dapp
