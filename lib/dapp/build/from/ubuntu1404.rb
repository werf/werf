module Dapp
  module Build
    module From
      module Ubuntu1404
        UBUNTU1404_COMMAND = ['apt-get update',
                              'apt-get -y dist-upgrade',
                              'apt-get -y install apt-utils git curl apt-transport-https git']
        UBUNTU1404_ENV     = %w(TERM='xterm' LANG='C.UTF-8')

        def ubuntu1404_signature
          hashsum [*UBUNTU1404_COMMAND, *UBUNTU1404_ENV]
        end

        def ubuntu1404_image
          DockerImage.new(from: DockerImage.new(name: 'ubuntu:14.04'), name: image_name).tap do |image|
            image.add_commands(UBUNTU1404_COMMAND)
            image.add_env(UBUNTU1404_ENV)
          end
        end
      end
    end
  end
end
