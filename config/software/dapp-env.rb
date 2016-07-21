name 'dapp-env'

dependency 'git'
dependency 'sudo'

build do
  link "#{install_dir}/embedded/bin", "#{install_dir}/bin"
end
