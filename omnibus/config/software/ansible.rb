name "ansible"

ANSIBLE_GIT_TAG = "v2.4.4.0+dapp-3"

dependency "python"
dependency "pip"

build do
  command "#{install_dir}/embedded/bin/pip install https://github.com/flant/ansible/archive/#{ANSIBLE_GIT_TAG}.tar.gz"
end
