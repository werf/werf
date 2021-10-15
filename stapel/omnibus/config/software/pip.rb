name "pip"

dependency "python"

build do
  command "curl https://bootstrap.pypa.io/pip/2.7/get-pip.py | #{install_dir}/embedded/bin/python"
end
