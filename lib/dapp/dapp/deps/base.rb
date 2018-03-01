module Dapp
  class Dapp
    module Deps
      module Base
        BASE_VERSION = "0.2.2"

        def base_container_name # FIXME: hashsum(image) or dockersafe()
          "dappdeps_base_#{BASE_VERSION}"
        end

        def base_container
          @base_container ||= begin
            is_container_exist = proc{shellout("#{host_docker} inspect #{base_container_name}").exitstatus.zero?}
            if !is_container_exist.call
              lock("dappdeps.container.#{base_container_name}", default_timeout: 600) do
                if !is_container_exist.call
                  log_secondary_process(t(code: 'process.base_container_creating', data: {name: base_container_name}), short: true) do
                    shellout!(
                      ["#{host_docker} create",
                      "--name #{base_container_name}",
                      "--volume /.dapp/deps/base/#{BASE_VERSION} dappdeps/base:#{BASE_VERSION}"].join(' ')
                    )
                  end
                end
              end
            end
            base_container_name
          end
        end

        def dappdeps_base_path
          "/.dapp/deps/base/#{BASE_VERSION}/embedded/bin:/.dapp/deps/base/#{BASE_VERSION}/embedded/sbin"
        end

        %w(rm rsync diff date cat
           stat readlink test sleep mkdir
           install sed cp true find
           bash tar sudo base64).each do |cmd|
          define_method("#{cmd}_bin") { "/.dapp/deps/base/#{BASE_VERSION}/embedded/bin/#{cmd}" }
        end

        def sudo_command(owner: nil, group: nil)
          sudo = ''
          if owner || group
            sudo = "#{sudo_bin} -E "
            sudo += "-u #{sudo_format_user(owner)} " if owner
            sudo += "-g #{sudo_format_user(group)} " if group
          end
          sudo
        end

        def sudo_format_user(user)
          user.to_s.to_i.to_s == user.to_s ? "\\\##{user}" : user
        end
      end # Base
    end # Deps
  end # Dapp
end # Dapp
