name "gnutls"

default_version "3.5.18"

version("3.5.18") { source md5: "c2d93d305ecbc55939bc2a8ed4a76a3d" }

license_file "LICENSE"

source url: "https://www.gnupg.org/ftp/gcrypt/gnutls/v#{version.split(".")[0,2].join(".")}/gnutls-#{version}.tar.xz"

relative_path "gnutls-#{version}"

dependency "libnettle"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command "./configure --prefix=#{install_dir}/embedded --without-p11-kit --with-included-libtasn1 --with-included-unistring", env: env
  command "make -j #{workers}", env: env
  command 'make install', env: env
end
