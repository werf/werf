apt_package 'htop'

cookbook_file '/app_install.txt' do
  source 'app_install/bar.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/bar.txt' do
  require 'securerandom'
  source 'app_install/bar.txt.erb'
  variables(var: SecureRandom.uuid)
end
