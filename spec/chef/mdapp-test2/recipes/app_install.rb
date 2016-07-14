apt_package 'jnettop'

cookbook_file "/#{cookbook_name.to_s.tr('-', '_')}_app_install.txt" do
  source 'app_install/koromyslo.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/koromyslo.txt' do
  require 'securerandom'
  source 'app_install/koromyslo.txt.erb'
  variables(var: SecureRandom.uuid)
end
