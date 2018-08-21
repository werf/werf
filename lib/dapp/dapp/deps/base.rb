module Dapp
  class Dapp
    module Deps
      module Base
        def base_container
          dappdeps_container(:base)
        end

        def dappdeps_base_path
          ruby2go_dappdeps_command(dappdeps: :base, command: :path)
        end

        %w(rm rsync diff date cat
           stat readlink test sleep mkdir
           install sed cp true find
           bash tar sudo base64).each do |bin|
          define_method("#{bin}_bin") { ruby2go_dappdeps_command(dappdeps: :base, command: :bin, options: { bin: bin }) }
        end

        def sudo_command(owner: nil, group: nil)
          ruby2go_dappdeps_command(dappdeps: :base, command: :sudo_command, options: { owner: owner, group: group })
        end
      end # Base
    end # Deps
  end # Dapp
end # Dapp
