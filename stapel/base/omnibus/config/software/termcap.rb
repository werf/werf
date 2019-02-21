name 'termcap'
default_version '1.3.1'

license 'GPL-2'
license_file 'doc/COPYING'

version('1.3.1') { source md5: 'ffe6f86e63a3a29fa53ac645faaabdfa' }

source url: "https://ftp.gnu.org/gnu/termcap/termcap-#{version}.tar.gz"

relative_path "termcap-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "./configure --prefix=#{install_dir}/embedded", env: env
  command "make -j #{workers}", env: env
  command 'make install', env: env
end
