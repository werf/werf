name 'sudo'
default_version '1.8.17p1'

license 'ISC'
license_file 'doc/LICENSE'

dependency 'zlib'

version '1.8.17p1' do
  source md5: '50a840a688ceb6fa3ab24fc0adf4fa23'
end

source url: "https://www.sudo.ws/dist/sudo-#{version}.tar.gz"

relative_path "sudo-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  command ["./configure CFLAGS=-D_PATH_SUDOERS=\\\\\\\"#{install_dir}/etc/sudoers\\\\\\\" ",
           "--prefix=#{install_dir}/embedded ",
           "--without-linux-audit --without-pam --without-secure-path"].join, env: env
  command "make -j #{workers}", env: env
  command 'make install', env: env

  mkdir "#{install_dir}/etc"
  command "echo 'root ALL=(ALL:ALL) ALL' > #{install_dir}/etc/sudoers"
end
