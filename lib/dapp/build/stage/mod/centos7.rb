module Dapp
  module Build
    module Stage # FIXME From
      module Mod # FIXME remove
        # Centos7 support
        module Centos7
          # rubocop:disable Metrics/MethodLength

# FIXME
#          CENTOS7_COMMAND =
#          CENTOS7_VOLUMES =
#          CENTOS7_ENV =
#
#          def centos7_signature
#            hashsum CENTOS7_COMMAND, CENTOS7_VOLUMES, CENTOS7_ENV
#          end

          def from_centos7(image)
            # FIXME image = DockerImage.new(from: "centos7"

            # use centos7
            image.add_commands('yum -y update; yum clean all',
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
                               'echo \'Defaults:root !requiretty\' >> /etc/sudoers')
            image.add_volume('/sys/fs/cgroup')
            image.add_env(%w(container='docker' TERM='xterm' LANG='en_US.UTF-8' LANGUAGE='en_US:en' LC_ALL='en_US.UTF-8'))

            image
          end
          # rubocop:enable Metrics/MethodLength
        end
      end
    end
  end
end
