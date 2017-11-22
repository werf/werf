name 'attr'
default_version '2.4.47'

license 'GPL-2'
license_file 'doc/COPYING'

version('2.4.47') { source md5: '84f58dec00b60f2dc8fd1c9709291cc7' }

source url: "http://download.savannah.gnu.org/releases/attr/attr-#{version}.src.tar.gz"

relative_path "attr-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "./configure --prefix=#{install_dir}/embedded", env: env
  command "make -j #{workers}", env: env
  command "make install", env: env
  command "make install-lib", env: env
  command "make install-dev", env: env
end
