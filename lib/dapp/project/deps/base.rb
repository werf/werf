module Dapp
  # Project
  class Project
    # Deps
    module Deps
      # Base
      module Base
        BASE_VERSION = '0.1.14'.freeze

        def base_container_name # FIXME: hashsum(image) or dockersafe()
          "dappdeps_base_#{BASE_VERSION}"
        end

        def base_container
          @base_container ||= begin
            if shellout("docker inspect #{base_container_name}").exitstatus.nonzero?
              log_secondary_process(t(code: 'process.base_container_creating'), short: true) do
                shellout!(
                  ['docker create',
                   "--name #{base_container_name}",
                   "--volume /.dapp/deps/base/#{BASE_VERSION} dappdeps/base:#{BASE_VERSION}"].join(' ')
                )
              end
            end
            base_container_name
          end
        end

        %w(rsync diff date cat
           stat sleep mkdir find
           install sed cp true
           bash tar sudo).each do |cmd|
          define_method("#{cmd}_bin") { "/.dapp/deps/base/#{BASE_VERSION}/bin/#{cmd}" }
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
  end # Project
end # Dapp
