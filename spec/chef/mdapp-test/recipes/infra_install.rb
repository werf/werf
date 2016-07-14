execute 'apt-get update' do
  command 'apt-get update'
end

apt_package 'sl'

cookbook_file "/#{cookbook_name.to_s.tr('-', '_')}_infra_install.txt" do
  source 'infra_install/pizza.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/pizza.txt' do
  require 'securerandom'
  source 'infra_install/pizza.txt.erb'
  variables(var: SecureRandom.uuid)
end
