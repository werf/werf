name "git"
default_version "2.19.2"

license "LGPL-2.1"
license_file "LGPL-2.1"
skip_transitive_dependency_licensing true

dependency "curl"
dependency "zlib"
dependency "openssl"
dependency "pcre"
#dependency "libiconv" # FIXME: can we figure out how to remove this?
dependency "expat"

relative_path "git-#{version}"

version "2.19.2" do
  source sha256: "db893ad69c9ac9498b09677c5839787eba2eb3b7ef2bc30bfba7e62e77cf7850"
end

version "2.17.1" do
  source sha256: "ec6452f0c8d5c1f3bcceabd7070b8a8a5eea11d4e2a04955c139b5065fd7d09a"
end

version "2.15.1" do
  source sha256: "85fca8781a83c96ba6db384cc1aa6a5ee1e344746bafac1cbe1f0fe6d1109c84"
end

version "2.14.1" do
  source sha256: "01925349b9683940e53a621ee48dd9d9ac3f9e59c079806b58321c2cf85a4464"
end

source url: "https://www.kernel.org/pub/software/scm/git/git-#{version}.tar.gz"

build do
  env = with_standard_compiler_flags(with_embedded_path)
  env["LDFLAGS"] += " -Wl,-rpath-link,/.werf/stapel/embedded/lib"

  env["EXTLIBS"] ||= ""
  env["EXTLIBS"] += "-ldl -lz"

  make "distclean"

  config_hash = {
    # Universal options
    NO_GETTEXT: "YesPlease",
    NEEDS_LIBICONV: "YesPlease",
    NO_INSTALL_HARDLINKS: "YesPlease",
    NO_PERL: "YesPlease",
    NO_PYTHON: "YesPlease",
    NO_TCLTK: "YesPlease",
    NO_OPENSSL: "YesPlease",
    NO_ICONV: "YesPlease",
    EXTLIBS: "-ldl -lz"
  }

  # Linux things!
  config_hash["HAVE_PATHS_H"] = "YesPlease"
  config_hash["NO_R_TO_GCC_LINKER"] = "YesPlease"

  erb source: "config.mak.erb",
    dest: "#{project_dir}/config.mak",
    mode: 0755,
    vars: {
            cc: env["CC"],
            ld: env["LD"],
            cflags: env["CFLAGS"],
            cppflags: env["CPPFLAGS"],
            install: env["INSTALL"],
            install_dir: install_dir,
            ldflags: env["LDFLAGS"],
            shell_path: env["SHELL_PATH"],
            config_hash: config_hash,
            }

  # NOTE - If you run ./configure the environment variables set above will not be
  # used and only the command line args will be used. The issue with this is you
  # cannot specify everything on the command line that you can with the env vars.
  make "prefix=#{install_dir}/embedded -j #{workers}", env: env
  make "prefix=#{install_dir}/embedded install", env: env
end
