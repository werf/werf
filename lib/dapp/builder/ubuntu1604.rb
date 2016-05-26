module Dapp
  module Builder
    # Ubuntu 16.04 support
    module Ubuntu1604
      def from_ubuntu1604
        # use ubuntu 16.04
        [
            'apt-get update',
            'apt-get -y dist-upgrade',
            'apt-get -y install apt-utils git curl apt-transport-https',
            {
                env: %w(TERM='xterm' LANG='C.UTF-8')
            }
        ]
      end
    end
  end
end
