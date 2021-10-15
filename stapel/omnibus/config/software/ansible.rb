name "ansible"

ANSIBLE_GIT_TAG = "v2.7.15"

dependency "libyaml"
dependency "python"
dependency "pip"

build do
  command "#{install_dir}/embedded/bin/pip install https://github.com/ansible/ansible/archive/#{ANSIBLE_GIT_TAG}.tar.gz"
  command "#{install_dir}/embedded/bin/pip install pyopenssl"
end
