module Dapp
  module Builder
    module Image
      # Ubuntu 16.04 support
      module Ubuntu1604
        def from_ubuntu1604
          # use ubuntu 16.04
          Dapp::Image.new(from: 'ubuntu:16.04').tap do |image|
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
