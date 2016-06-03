module Dapp
  module Builder
    module Mod
      # Ubuntu 14.04 support
      module Ubuntu1404
        def from_ubuntu1404
          # use ubuntu 14.04
          Dapp::Image.new(from: 'ubuntu:14.04').tap do |image|
            image.build_cmd!('apt-get update',
                             'apt-get -y dist-upgrade',
                             'apt-get -y install apt-utils git curl apt-transport-https')

            image.build_opts!({ env: %w(TERM='xterm' LANG='C.UTF-8') })
          end
        end
      end
    end
  end
end
