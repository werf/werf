name 'dapp-env'

license 'MIT'
license_file 'https://github.com/flant/dapp-env/blob/master/LICENSE.txt'

dependency 'git'
dependency 'sudo'

build do
  link "#{install_dir}/embedded/bin", "#{install_dir}/bin"
end
