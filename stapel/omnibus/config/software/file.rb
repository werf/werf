
name "file"
default_version "5.36"

license_file "COPYING"

version("5.36") { source md5: "9af0eb3f5db4ae00fffc37f7b861575c" }

source url: "ftp://ftp.astron.com/pub/file/file-#{version}.tar.gz"

relative_path "file-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "./configure --prefix=#{install_dir}/embedded", env: env
  command "make -j #{workers}", env: env
  command "make install", env: env
end
