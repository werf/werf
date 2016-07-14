apt_package 'automake'

cookbook_file "/#{cookbook_name.to_s.tr('-', '_')}_app_setup.txt" do
  source 'app_setup/pelmeni.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/pelmeni.txt' do
  require 'securerandom'
  source 'app_setup/pelmeni.txt.erb'
  variables(var: SecureRandom.uuid)
end
