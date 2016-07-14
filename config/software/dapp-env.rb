name 'dapp-env'

dependency 'git'

build do
  link "#{install_dir}/embedded/bin", "#{install_dir}/bin"
end
