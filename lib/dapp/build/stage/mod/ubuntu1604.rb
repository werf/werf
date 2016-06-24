module Dapp
  module Build
    module Stage
      module Mod
        # Ubuntu 16.04 support
        module Ubuntu1604
          def from_ubuntu1604(image)
            # use ubuntu 16.04
            image.add_commands('apt-get update',
                               'apt-get -y dist-upgrade',
                               'apt-get -y install apt-utils git curl apt-transport-https git')
            image.add_env(%w(TERM='xterm' LANG='C.UTF-8'))
            image
          end
        end
      end
    end
  end
end
