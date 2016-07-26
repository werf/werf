package 'cmake'

cookbook_file "/#{cookbook_name.to_s.tr('-', '_')}_app_setup.txt" do
  source 'app_setup/taburetka.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/taburetka.txt' do
  require 'securerandom'
  source 'app_setup/taburetka.txt.erb'
  variables(var: SecureRandom.uuid)
end
