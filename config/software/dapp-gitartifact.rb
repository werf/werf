name 'dapp-gitartifact'

license 'MIT'
license_file 'https://github.com/flant/dapp-gitartifact/blob/master/LICENSE.txt'

dependency 'git'
dependency 'sudo'

build do
  link "#{install_dir}/embedded/bin", "#{install_dir}/bin"
end
