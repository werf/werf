name "shadow"

default_version "4.5"

license_file 'COPYING'

version("4.5") { source md5: "c350da50c2120de6bb29177699d89fe3" }

source url: "https://github.com/shadow-maint/shadow/releases/download/#{version}/shadow-#{version}.tar.xz"

relative_path "shadow-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "./configure --prefix=#{install_dir}/embedded", env: env
  command "make -j #{workers}", env: env
  command 'make install', env: env
end
