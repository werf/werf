module Dapp
  module Builder
    # Ubuntu 16.04 support
    module Ubuntu1404
      def from_ubuntu1404
        # use ubuntu 14.04
        [
            'apt-get update',
            'apt-get -y dist-upgrade',
            'apt-get -y install apt-utils git curl apt-transport-https',
            {
                env: ["TERM='xterm'", "LANG='C.UTF-8'"]
            }
        ]
      end
    end
  end
end
