module Dapp
  class Dapp
    module Deps
      module Toolchain
        TOOLCHAIN_VERSION = "0.1.1"

        def toolchain_container_name # FIXME: hashsum(image) or dockersafe()
          "dappdeps_toolchain_#{TOOLCHAIN_VERSION}"
        end

        def toolchain_container
          @toolchain_container ||= begin
            is_container_exist = proc {shellout("#{host_docker} inspect #{toolchain_container_name}").exitstatus.zero?}
            if !is_container_exist.call
              lock("dappdeps.container.#{toolchain_container_name}", default_timeout: 300, type: :global) do
                if !is_container_exist.call
                  log_secondary_process(t(code: 'process.toolchain_container_creating', data: {name: toolchain_container_name}), short: true) do
                    shellout!(
                      ["#{host_docker} create",
                      "--name #{toolchain_container_name}",
                      "--volume /.dapp/deps/toolchain/#{TOOLCHAIN_VERSION} dappdeps/toolchain:#{TOOLCHAIN_VERSION}"].join(' ')
                    )
                  end
                end
              end
            end
            toolchain_container_name
          end
        end
      end # Toolchain
    end # Deps
  end # Dapp
end # Dapp
