module Dapp
  module Build
    module From
      module Centos7
        CENTOS7_COMMAND = ['yum -y update; yum clean all',
                           '(cd /lib/systemd/system/sysinit.target.wants/; for i in *; do [ $i == systemd-tmpfiles-setup.service ] || rm -f $i; done)',
                           'rm -f /lib/systemd/system/multi-user.target.wants/*',
                           'rm -f /etc/systemd/system/*.wants/*',
                           'rm -f /lib/systemd/system/local-fs.target.wants/*',
                           'rm -f /lib/systemd/system/sockets.target.wants/*udev*',
                           'rm -f /lib/systemd/system/sockets.target.wants/*initctl*',
                           'rm -f /lib/systemd/system/basic.target.wants/*',
                           'rm -f /lib/systemd/system/anaconda.target.wants/*',
                           'rm -f /lib/systemd/system/systemd-remount-fs.service',
                           'rm -f /lib/systemd/system/sys-fs-fuse-connections.mount',
                           '/usr/sbin/init', # add real systemd
                           'yum makecache', # cache yum fastestmirror
                           'localedef -c -f UTF-8 -i en_US en_US.UTF-8', # TERM & utf8
                           'sed \'s/\(-\?session\s\+optional\s\+pam_systemd\.so.*\)/#\1/g\' -i /etc/pam.d/system-auth', # centos hacks
                           'yum install -y sudo git patch',
                           'echo \'Defaults:root !requiretty\' >> /etc/sudoers']
        CENTOS7_VOLUMES =  ['/sys/fs/cgroup']
        CENTOS7_ENV     =  %w(container='docker' TERM='xterm' LANG='en_US.UTF-8' LANGUAGE='en_US:en' LC_ALL='en_US.UTF-8')

        def centos7_signature
          hashsum ['centos7', CENTOS7_COMMAND, CENTOS7_VOLUMES, CENTOS7_ENV]
        end

        def centos7_image
          DockerImage.new(from: DockerImage.new(name: 'centos:7'), name: image_name).tap do |image|
            image.add_commands(CENTOS7_COMMAND)
            image.add_volume(CENTOS7_VOLUMES)
            image.add_env(CENTOS7_ENV)
          end
        end
      end
    end
  end
end
