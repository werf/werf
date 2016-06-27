module Dapp
  module Build
    module From
      module Ubuntu1604
        UBUNTU1604_COMMAND = ['apt-get update',
                              'apt-get -y dist-upgrade',
                              'apt-get -y install apt-utils git curl apt-transport-https git']
        UBUNTU1604_ENV     = %w(TERM='xterm' LANG='C.UTF-8')

        def ubuntu1604_signature
          hashsum [*UBUNTU1604_COMMAND, *UBUNTU1604_ENV]
        end

        def ubuntu1604_image
          DockerImage.new(from: DockerImage.new(name: 'ubuntu:16.04'), name: image_name).tap do |image|
            image.add_commands(UBUNTU1604_COMMAND)
            image.add_env(UBUNTU1604_ENV)
          end
        end
      end
    end
  end
end
