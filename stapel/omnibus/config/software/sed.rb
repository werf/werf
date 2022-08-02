name 'sed'
default_version '4.2'

license 'GPL-3.0'
license_file 'COPYING'

version('4.2') { source md5: '31580bee0c109c0fc8f31c4cf204757e' }

source url: "https://ftp.gnu.org/gnu/sed/sed-#{version}.tar.gz"

dependency 'attr'
dependency 'acl'

relative_path "sed-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "./configure --prefix=#{install_dir}/embedded", env: env
  command "make -j #{workers}", env: env
  command 'make install', env: env
end
