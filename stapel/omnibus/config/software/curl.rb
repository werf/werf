#
# Copyright:: Chef Software, Inc.
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

name "curl"
default_version "7.79.1"

dependency "zlib"
dependency "openssl"
dependency "cacerts"

license "MIT"
license_file "COPYING"
skip_transitive_dependency_licensing true

# version_list: url=https://curl.se/download/ filter=*.tar.gz
version("7.79.1") { source sha256: "370b11201349816287fb0ccc995e420277fbfcaf76206e309b3f60f0eda090c2" }
version("7.79.0") { source sha256: "aff0c7c4a526d7ecc429d2f96263a85fa73e709877054d593d8af3d136858074" }
version("7.78.0") { source sha256: "ed936c0b02c06d42cf84b39dd12bb14b62d77c7c4e875ade022280df5dcc81d7" }
version("7.77.0") { source sha256: "b0a3428acb60fa59044c4d0baae4e4fc09ae9af1d8a3aa84b2e3fbcd99841f77" }
version("7.76.1") { source sha256: "5f85c4d891ccb14d6c3c701da3010c91c6570c3419391d485d95235253d837d7" }

source url: "https://curl.haxx.se/download/curl-#{version}.tar.gz"

relative_path "curl-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  if freebsd?
    # from freebsd ports - IPv6 Hostcheck patch
    patch source: "curl-freebsd-hostcheck.patch", plevel: 1, env: env
  end

  delete "#{project_dir}/src/tool_hugehelp.c"

  if aix?
    # alpn doesn't appear to work on AIX when connecting to certain sites, most
    # importantly for us https://www.github.com Since git uses libcurl under
    # the covers, this functionality breaks the handshake on connection, giving
    # a cryptic error. This patch essentially forces disabling of ALPN on AIX,
    # which is not really what we want in a http/2 world, but we're not there
    # yet.
    patch_env = env.dup
    patch_env["PATH"] = "/opt/freeware/bin:#{env["PATH"]}" if aix?
    patch source: "curl-aix-disable-alpn.patch", plevel: 0, env: patch_env

    # otherwise gawk will die during ./configure with variations on the theme of:
    # "/opt/omnibus-toolchain/embedded/lib/libiconv.a(shr4.o) could not be loaded"
    env["LIBPATH"] = "/usr/lib:/lib"
  elsif solaris2?
    # Without /usr/gnu/bin first in PATH the libtool fails during make on Solaris
    env["PATH"] = "/usr/gnu/bin:#{env["PATH"]}"
  end

  env["LIBS"] ||= ""
  env["LIBS"] += "-ldl -lz"

  configure_options = [
    "--prefix=#{install_dir}/embedded",
    "--disable-option-checking",
    "--disable-manual",
    "--disable-debug",
    "--enable-optimize",
    "--disable-ldap",
    "--disable-ldaps",
    "--disable-rtsp",
    "--enable-proxy",
    "--disable-pop3",
    "--disable-imap",
    "--disable-smtp",
    "--disable-gopher",
    "--disable-dependency-tracking",
    "--enable-ipv6",
    "--without-libidn2",
    "--without-gnutls",
    "--without-librtmp",
    "--without-zsh-functions-dir",
    "--without-fish-functions-dir",
    "--disable-mqtt",
    "--with-ssl=#{install_dir}/embedded",
    #"--without-ssl",
    "--with-zlib=#{install_dir}/embedded",
    "--with-ca-bundle=#{install_dir}/embedded/ssl/certs/cacert.pem",
    "--without-zstd",
  ]

  configure(*configure_options, env: env)

  make "-j #{workers}", env: env
  make "install", env: env
end
