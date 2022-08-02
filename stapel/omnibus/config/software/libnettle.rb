name "libnettle"

default_version "3.4"

version("3.4") { source md5: "dc0f13028264992f58e67b4e8915f53d" }

license_file "COPYINGv2"

source url: "https://ftp.gnu.org/gnu/nettle/nettle-#{version}.tar.gz"

relative_path "nettle-#{version}"

dependency "gmp"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "./configure --prefix=#{install_dir}/embedded", env: env
  command "make -j #{workers}", env: env
  command 'make install', env: env
end
