#
# Copyright 2016 Chef Software, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

name "gtar"
default_version "1.29"

version("1.29") { source md5: "c57bd3e50e43151442c1995f6236b6e9" }
version("1.28") { source md5: "6ea3dbea1f2b0409b234048e021a9fd7" }

license "GPL-3.0"
license_file "COPYING"

source url: "http://ftp.gnu.org/gnu/tar/tar-#{version}.tar.gz"

relative_path "tar-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  configure_command = [
    "./configure",
    "--prefix=#{install_dir}/embedded",
  ]

  configure_command << "--without-posix-acls"
  configure_command << "--without-xattrs"
  env["FORCE_UNSAFE_CONFIGURE"] = "1"

  # First off let's disable selinux support, as it causes issues on some platforms
  # We're not doing it on every platform because this breaks on OSX
  unless osx?
    configure_command << " --without-selinux"
  end

  if s390x?
    # s390x doesn't support posix acls
    configure_command << " --without-posix-acls"
  elsif aix? && version.satisfies?("< 1.32")
    # xlc doesn't allow duplicate entries in case statements
    patch_env = env.dup
    patch_env["PATH"] = "/opt/freeware/bin:#{env["PATH"]}"
    patch source: "aix_extra_case.patch", plevel: 0, env: patch_env
  end

  command configure_command.join(" "), env: env
  make "-j #{workers}", env: env
  make "-j #{workers} install", env: env
end
