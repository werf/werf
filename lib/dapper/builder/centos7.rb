module Dapper
  class Builder
    # Centos7 support
    module Centos7
      # rubocop:disable Metrics/MethodLength:
      def from_centos7
        # use centos7
        docker.from 'centos:7'

        # add real systemd
        docker.env container: 'docker', step: :begining
        docker.run 'yum -y swap -- remove systemd-container systemd-container-libs -- install systemd systemd-libs', step: :begining
        docker.run(
          'yum -y update; yum clean all',
          '(cd /lib/systemd/system/sysinit.target.wants/; for i in *; do [ $i == systemd-tmpfiles-setup.service ] || rm -f $i; done)',
          'rm -f /lib/systemd/system/multi-user.target.wants/*',
          'rm -f /etc/systemd/system/*.wants/*',
          'rm -f /lib/systemd/system/local-fs.target.wants/*',
          'rm -f /lib/systemd/system/sockets.target.wants/*udev*',
          'rm -f /lib/systemd/system/sockets.target.wants/*initctl*',
          'rm -f /lib/systemd/system/basic.target.wants/*',
          'rm -f /lib/systemd/system/anaconda.target.wants/*',
          step: :begining
        )
        docker.volume '/sys/fs/cgroup', step: :begining
        docker.cmd '/usr/sbin/init', step: :begining

        # cache yum fastestmirror
        docker.run 'yum makecache', step: :begining

        # TERM & utf8
        docker.run 'localedef -c -f UTF-8 -i en_US en_US.UTF-8', step: :begining
        docker.env TERM: 'xterm', LANG: 'en_US.UTF-8', LANGUAGE: 'en_US:en', LC_ALL: 'en_US.UTF-8', step: :begining

        # centos hacks
        docker.run(
          'sed \'s/\(-\?session\s\+optional\s\+pam_systemd\.so.*\)/#\1/g\' -i /etc/pam.d/system-auth',
          'yum install -y sudo git',
          'echo \'Defaults:root !requiretty\' >> /etc/sudoers',
          step: :begining
        )
      end
      # rubocop:enable Metrics/MethodLength
    end
  end
end
