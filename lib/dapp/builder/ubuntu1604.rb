module Dapp
  class Builder
    # Ubuntu 16.04 support
    module Ubuntu1604
      def from_ubuntu1604
        # use ubuntu 16.04
        docker.from 'ubuntu:16.04'
        docker.run(
          'apt-get update',
          'apt-get dist-upgrade',
          'apt-get -y install apt-utils git curl',
          step: :begining
        )

        # TERM & utf8
        docker.env TERM: 'xterm', LANG: 'C.UTF-8', step: :begining
      end
    end
  end
end
