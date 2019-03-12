name 'coreutils'
default_version '8.30'

license 'GPL-3.0'
license_file 'COPYING'

version('8.25') { source md5: '070e43ba7f618d747414ef56ab248a48' }
version('8.30') { source md5: 'ab06d68949758971fe744db66b572816' }

source url: "https://ftp.gnu.org/gnu/coreutils/coreutils-#{version}.tar.xz"

dependency 'pcre'
dependency 'acl'
dependency 'gmp'

relative_path "coreutils-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command [
    'FORCE_UNSAFE_CONFIGURE=1 ./configure',
    "--prefix=#{install_dir}/embedded",
    '--without-selinux'
  ].join(' '), env: env

  command "make -j #{workers}", env: env
  command 'make install', env: env
end
