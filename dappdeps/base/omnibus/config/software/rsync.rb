name 'rsync'
default_version '3.1.2'

license 'GPL-3.0'
license_file 'COPYING'

version('3.1.2') { source md5: '0f758d7e000c0f7f7d3792610fad70cb' }

source url: "https://download.samba.org/pub/rsync/src/rsync-#{version}.tar.gz"

dependency 'attr'
dependency 'acl'
dependency 'popt'

relative_path "rsync-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "./configure --prefix=#{install_dir}/embedded", env: env
  command "make -j #{workers}", env: env
  command 'make install', env: env
end
