name "e2fsprogs"
default_version "1.45.0"

license "GPL-2.0"
license_file "ext2ed/COPYRIGHT"

version("1.45.0") { source md5: "641e1371dbdd118eade96bb963104f16" }

source url: "https://datapacket.dl.sourceforge.net/project/e2fsprogs/e2fsprogs/v#{version}/e2fsprogs-#{version}.tar.gz"

relative_path "e2fsprogs-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "./configure --prefix=#{install_dir}/embedded", env: env
  command "make -j #{workers}", env: env
  command "make install", env: env
end
