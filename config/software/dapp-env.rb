name 'dapp-env'

license 'MIT'
license_file '../LICENSE.txt'

dependency 'git'
dependency 'sudo'

build do
  link "#{install_dir}/embedded/bin", "#{install_dir}/bin"
end
