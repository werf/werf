name 'libaudit'
default_version '2.4.2'

dependency 'libffi'

version '2.4.2' do
  source md5: '2874185ee03af01bce47d00d66c291d5'
end

source url: "https://github.com/Distrotech/libaudit/archive/audit-#{version}.tar.gz"

relative_path "libaudit-audit-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)
  command './autogen.sh', env: env
  command "./configure --disable-listener --prefix=#{install_dir}/embedded", env: env
  command "make -j #{workers}", env: env
  command 'make install', env: env
end
