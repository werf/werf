apt_package 'gcc'

cookbook_file "/#{cookbook_name}.infra_setup.txt" do
  source 'infra_setup/burger.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/burger.txt' do
  require 'securerandom'
  source 'infra_setup/burger.txt.erb'
  variables(var: SecureRandom.uuid)
end
