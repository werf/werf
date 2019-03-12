name 'findutils'
default_version '4.6.0'

license 'GPL-3.0'
license_file 'COPYING'

version('4.6.0') { source md5: '9936aa8009438ce185bea2694a997fc1' }

source url: "https://ftp.gnu.org/pub/gnu/findutils/findutils-#{version}.tar.gz"

dependency 'pcre'

relative_path "findutils-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "sed -i 's/IO_ftrylockfile/IO_EOF_SEEN/' gl/lib/*.c"
  command "sed -i '/unistd/a #include <sys/sysmacros.h>' gl/lib/mountlist.c"
  command "echo \"#define _IO_IN_BACKUP 0x100\" >> gl/lib/stdio-impl.h"

  command "./configure --prefix=#{install_dir}/embedded --without-selinux", env: env
  command "make -j #{workers}", env: env
  command 'make install', env: env
end
