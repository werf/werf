name "python-apt"

dependency "python"
dependency "python-distutils-extra"

build do
  command "git clone https://anonscm.debian.org/cgit/apt/python-apt.git"
  command "cd python-apt && #{install_dir}/embedded/bin/python setup.py install"
end
