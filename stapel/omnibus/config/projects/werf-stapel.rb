name "werf-stapel"
maintainer "Flant"
homepage "https://github.com/werf/werf"

license "Apache 2.0"
license_file "LICENSE"

install_dir "/.werf/stapel"

build_version "1"
build_iteration 1

dependency "werf-stapel"

orig_rm_rf = FileUtils.method(:rm_rf)
FileUtils.define_singleton_method(:rm_rf) do |list, **kwargs|
  return if list == "/.werf/stapel"
  return orig_rm_rf.call(list, kwargs)
end
