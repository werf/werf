include_recipe 'apt' if node[:platform_family].to_s == 'debian'

cookbook_file "/#{cookbook_name.to_s.tr('-', '_')}_before_install.txt" do
  source 'batareika.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/batareika.txt' do
  require 'securerandom'
  source 'batareika.txt.erb'
  variables(var: SecureRandom.uuid)
end
