name "libapt"

default_version "1.5.1"

version("1.5.1") { source md5: "3b501e13f4441b065e7da9bda641e77e" }

license_file "COPYING"

source url: "https://github.com/Debian/apt/archive/#{version}.tar.gz"

relative_path "apt-#{version}"

dependency "berkeley-db"
dependency "curl"
dependency "gnutls"
dependency "bzip2"
dependency "liblzma"
dependency "liblz4"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "cmake -DCMAKE_INSTALL_PREFIX=#{install_dir}/embedded", env: env

  command "cd apt-pkg && make -j #{workers}", env: env
  command "cd apt-pkg && make install", env: env

  command "cd apt-inst && make -j #{workers}", env: env
  command "cd apt-inst && make install", env: env
end
