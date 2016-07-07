apt_package 'unzip'

cookbook_file "/#{cookbook_name.to_s.gsub('-', '_')}_app_setup.txt" do
  source 'app_setup/qux.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/qux.txt' do
  require 'securerandom'
  source 'app_setup/qux.txt.erb'
  variables(var: SecureRandom.uuid)
end
