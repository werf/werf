name "sqlite3"
default_version "3.29.0"

version("3.29.0") { source md5: "8f3dfe83387e62ecb91c7c5c09c688dc" }

license "-"
license_file "-"

source url: "https://sqlite.org/2019/sqlite-autoconf-3290000.tar.gz"

relative_path "sqlite-autoconf-3290000"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "./configure --prefix=#{install_dir}/embedded", env: env
  command "make -j #{workers}", env: env
  command 'make install', env: env
end
