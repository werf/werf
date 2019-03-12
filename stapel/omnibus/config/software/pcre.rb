name "pcre"
default_version "8.38"

license "BSD-2-Clause"
license_file "LICENCE"
skip_transitive_dependency_licensing true

dependency "libedit"
dependency "ncurses"
dependency "config_guess"

version "8.38" do
  source md5: "8a353fe1450216b6655dfcf3561716d9"
end

version "8.31" do
  source md5: "fab1bb3b91a4c35398263a5c1e0858c1"
end

source url: "http://downloads.sourceforge.net/project/pcre/pcre/#{version}/pcre-#{version}.tar.gz"

relative_path "pcre-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  update_config_guess

  command "./configure" \
          " --prefix=#{install_dir}/embedded" \
          " --disable-cpp" \
          " --enable-utf" \
          " --enable-unicode-properties", env: env

  make "-j #{workers}", env: env
  make "install", env: env
end
