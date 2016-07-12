apt_package 'sl'

cookbook_file "/#{cookbook_name.to_s.gsub('-', '_')}_infra_install.txt" do
  source 'infra_install/batareika.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/batareika.txt' do
  require 'securerandom'
  source 'infra_install/batareika.txt.erb'
  variables(var: SecureRandom.uuid)
end
