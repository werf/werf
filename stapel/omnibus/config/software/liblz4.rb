name "liblz4"

default_version "1.8.1.2"

version("1.8.1.2") { source md5: "343538e69ba752a386c669b1a28111e2" }

license_file "LICENSE"

source url: "https://github.com/lz4/lz4/archive/v#{version}.tar.gz"

relative_path "lz4-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "cd lib && make", env: env
  command "cd lib && make install PREFIX=#{install_dir}/embedded", env: env
end
