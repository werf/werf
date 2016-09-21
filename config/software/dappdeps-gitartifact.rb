name 'dappdeps-gitartifact'

license 'MIT'
license_file 'https://github.com/flant/dappdeps-gitartifact/blob/master/LICENSE.txt'

dependency 'git'
dependency 'sudo'

build do
  link "#{install_dir}/embedded/bin", "#{install_dir}/bin"
end
