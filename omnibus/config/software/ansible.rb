name "ansible"

ANSIBLE_GIT_TAG = "v2.4.1.0-1"

dependency "python"
dependency "pip"

build do
  command "#{install_dir}/embedded/bin/pip install https://github.com/ansible/ansible/archive/#{ANSIBLE_GIT_TAG}.tar.gz"
end
