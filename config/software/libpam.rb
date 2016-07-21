name 'libpam'
default_version '1.3.0'

version '1.3.0' do
  source md5: '642c994edb9cf39b74cb1bca0649613b'
end

source url: "http://www.linux-pam.org/library/Linux-PAM-#{version}.tar.gz"

relative_path "Linux-PAM-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)
  command "./configure --prefix=#{install_dir}/embedded", env: env
  command "make -j #{workers}", env: env
  command 'make install', env: env
end
