apt_package 'make'

cookbook_file "/#{cookbook_name.to_s.gsub('-', '_')}_app_install.txt" do
  source 'app_install/taco.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/taco.txt' do
  require 'securerandom'
  source 'app_install/taco.txt.erb'
  variables(var: SecureRandom.uuid)
end
