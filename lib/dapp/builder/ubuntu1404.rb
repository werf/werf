module Dapp
  class Builder
    # Ubuntu 16.04 support
    module Ubuntu1404
      def from_ubuntu1404
        # use ubuntu 14.04
        docker.from 'ubuntu:14.04'
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
