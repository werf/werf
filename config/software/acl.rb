name 'acl'
default_version '2.2.52'

license 'GPL-2'
license_file 'doc/COPYING'

version('2.2.52') { source md5: 'a61415312426e9c2212bd7dc7929abda' }

source url: "http://download.savannah.gnu.org/releases/acl/acl-#{version}.src.tar.gz"

dependency 'attr'

relative_path "acl-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "./configure --prefix=#{install_dir}/embedded", env: env
  command "make -j #{workers}", env: env
  command 'make install-lib', env: env
  command 'make install', env: env
end
