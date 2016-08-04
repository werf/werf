#package 'unzip'

cookbook_file "/#{cookbook_name.to_s.tr('-', '_')}_app_setup.txt" do
  source 'qux.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/qux.txt' do
  require 'securerandom'
  source 'qux.txt.erb'
  variables(var: SecureRandom.uuid)
end
