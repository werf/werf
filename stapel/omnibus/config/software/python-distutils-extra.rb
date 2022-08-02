name "python-distutils-extra"

dependency "python"

default_version "2.39"

version("2.39") { source md5: "16e06db0ef73a35b4bff4b9eed5699b5" }

license_file "LICENSE"

source url: "https://launchpad.net/python-distutils-extra/trunk/#{version}/+download/python-distutils-extra-#{version}.tar.gz"

relative_path "python-distutils-extra-#{version}"

build do
  command "#{install_dir}/embedded/bin/python setup.py build"
  command "#{install_dir}/embedded/bin/python setup.py install"
end
