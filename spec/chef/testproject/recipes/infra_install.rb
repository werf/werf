execute 'apt-get update' do
  command 'apt-get update'
end

apt_package 'curl'

cookbook_file "/#{cookbook_name}.infra_install.txt" do
  source 'infra_install/foo.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/foo.txt' do
  require 'securerandom'
  source 'infra_install/foo.txt.erb'
  variables(var: SecureRandom.uuid)
end
