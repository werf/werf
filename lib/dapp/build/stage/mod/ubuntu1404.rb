module Dapp
  module Build
    module Stage
      module Mod
        # Ubuntu 14.04 support
        module Ubuntu1404
          def from_ubuntu1404(image)
            # use ubuntu 14.04
            image.build_cmd!('apt-get update',
                             'apt-get -y dist-upgrade',
                             'apt-get -y install apt-utils git curl apt-transport-https git')

            image.build_opts!({ env: %w(TERM='xterm' LANG='C.UTF-8') })
            image
          end
        end
      end
    end
  end
end
