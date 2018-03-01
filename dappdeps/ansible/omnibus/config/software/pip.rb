name "pip"

dependency "python"

build do
  command "curl https://bootstrap.pypa.io/get-pip.py | #{install_dir}/embedded/bin/python"
end
