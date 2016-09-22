name 'diffutils'
default_version '3.5'

license 'GPL-3.0'
license_file 'COPYING'

version('3.5') { source md5: '569354697ff1cfc9a9de3781361015fa' }

source url: "https://ftp.gnu.org/gnu/diffutils/diffutils-#{version}.tar.xz"

relative_path "diffutils-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "./configure --prefix=#{install_dir}/embedded", env: env
  command "make -j #{workers}", env: env
  command 'make install', env: env
end
