name "berkeley-db"

default_version "6.2.32"

version("6.2.32") { source md5: "b24a25b1c84f8baaa8de7f634d8d3cad" }

license_file "LICENSE"

source url: "http://download.oracle.com/berkeley-db/db-#{version}.NC.tar.gz"

relative_path "db-#{version}.NC"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "cd build_unix && ../dist/configure --prefix=#{install_dir}/embedded", env: env
  command "cd build_unix && make -j #{workers}", env: env
  command "cd build_unix && make install", env: env
end
