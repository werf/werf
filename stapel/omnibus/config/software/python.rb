#
# Copyright 2013-2015 Chef Software, Inc.
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

name "python"
default_version "2.7.18"

license "Python-2.0"
license_file "LICENSE"
skip_transitive_dependency_licensing true

dependency "ncurses"
dependency "zlib"
dependency "openssl"
dependency "bzip2"

# version_list: url=https://www.python.org/ftp/python/#{version}/ filter=*.tgz

version("2.7.18") { source sha256: "da3080e3b488f648a3d7a4560ddee895284c3380b11d6de75edb986526b9a814" }
version("2.7.14") { source sha256: "304c9b202ea6fbd0a4a8e0ad3733715fbd4749f2204a9173a58ec53c32ea73e8" }
version("2.7.9")  { source sha256: "c8bba33e66ac3201dabdc556f0ea7cfe6ac11946ec32d357c4c6f9b018c12c5b" }
version("2.7.5")  { source sha256: "8e1b5fa87b91835afb376a9c0d319d41feca07ffebc0288d97ab08d64f48afbf" }

source url: "https://python.org/ftp/python/#{version}/Python-#{version}.tgz"

relative_path "Python-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  if mac_os_x?
    os_x_release = ohai["platform_version"].match(/([0-9]+\.[0-9]+).*/).captures[0]
    env["MACOSX_DEPLOYMENT_TARGET"] = os_x_release
  end

  command "./configure" \
          " --prefix=#{install_dir}/embedded" \
          " --enable-shared" \
          " --with-dbmliborder=", env: env

  command "sed -i -e 's|os.unlink(tmpfile)|pass|' ./setup.py"

  make env: env
  make "install", env: env

  # There exists no configure flag to tell Python to not compile readline
  delete "#{install_dir}/embedded/lib/python2.7/lib-dynload/readline.*"

  # Ditto for sqlite3
  delete "#{install_dir}/embedded/lib/python2.7/lib-dynload/_sqlite3.*"
  delete "#{install_dir}/embedded/lib/python2.7/sqlite3/"

  # Remove unused extension which is known to make healthchecks fail on CentOS 6
  delete "#{install_dir}/embedded/lib/python2.7/lib-dynload/_bsddb.*"

  # Remove sqlite3 libraries, if you want to include sqlite, create a new def
  # in your software project and build it explicitly. This removes the adapter
  # library from python, which links incorrectly to a system library. Adding
  # your own sqlite definition will fix this.
  delete "#{install_dir}/embedded/lib/python2.7/lib-dynload/_sqlite3.*"
end
