name 'glibc'
default_version '2.24'

version('2.24') { source md5: '97dc5517f92016f3d70d83e3162ad318' }

license 'GPL-2.0'
license_file 'COPYING'

dependency 'pcre'

source url: "https://ftp.gnu.org/gnu/glibc/glibc-#{version}.tar.xz"

relative_path "glibc-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  build_dir = "../glibc-#{version}-build"

  command([
    'export SRC_DIR=$(pwd)',
    "mkdir #{build_dir}",
    "cd #{build_dir}",
    ['$SRC_DIR/configure',
     "--prefix=#{install_dir}/embedded",
     '--enable-static-nss',
     '--disable-nss-crypt',
     '--disable-build-nscd',
     '--disable-nscd',
     '--without-selinux',
     'libc_cv_forced_unwind=yes',
     'libc_cv_c_cleanup=yes'].join(' ')
  ].join(' && '), env: env)

  command "cd #{build_dir} && make -j #{workers}", env: env
  command "cd #{build_dir} && make install", env: env
end
