module Dapp
  # Project
  class Project
    # Deps
    module Deps
      # Base
      module Base
        BASE_IMAGE = 'dappdeps/base:0.1.13'.freeze

        def base_container_name # FIXME: hashsum(image) or dockersafe()
          BASE_IMAGE.tr('/', '_').tr(':', '_')
        end

        def base_container
          @base_container ||= begin
            if shellout("docker inspect #{base_container_name}").exitstatus.nonzero?
              log_secondary_process(t(code: 'process.base_container_loading'), short: true) do
                shellout!(
                  ['docker create',
                   "--name #{base_container_name}",
                   "--volume /.dapp/deps/base #{BASE_IMAGE}"].join(' ')
                )
              end
            end
            base_container_name
          end
        end

        def rsync_path
          '/.dapp/deps/base/bin/rsync'
        end

        def diff_path
          '/.dapp/deps/base/bin/diff'
        end

        def date_path
          '/.dapp/deps/base/bin/date'
        end

        def echo_path
          '/.dapp/deps/base/bin/echo'
        end

        def stat_path
          '/.dapp/deps/base/bin/stat'
        end

        def sleep_path
          '/.dapp/deps/base/bin/sleep'
        end

        def mkdir_path
          '/.dapp/deps/base/bin/mkdir'
        end

        def find_path
          '/.dapp/deps/base/bin/find'
        end

        def install_path
          '/.dapp/deps/base/bin/install'
        end

        def sed_path
          '/.dapp/deps/base/bin/sed'
        end

        def cp_path
          '/.dapp/deps/base/bin/cp'
        end

        def true_path
          '/.dapp/deps/base/bin/true'
        end

        def bash_path
          '/.dapp/deps/base/bin/bash'
        end

        def tar_path
          '/.dapp/deps/base/bin/tar'
        end

        def sudo_path
          '/.dapp/deps/base/bin/sudo'
        end

        def sudo_command(owner: nil, group: nil)
          sudo = ''
          if owner || group
            sudo = "#{sudo_path} -E "
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
  end # Project
end # Dapp
