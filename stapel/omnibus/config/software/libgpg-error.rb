name "libgpg-error"
default_version "1.36"

license "GPL-2"
license_file "COPYING"

version("1.36") { source md5: "eff437f397e858a9127b76c0d87fa5ed" }

source url: "https://gnupg.org/ftp/gcrypt/libgpg-error/libgpg-error-#{version}.tar.bz2"

relative_path "libgpg-error-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command ""
end
