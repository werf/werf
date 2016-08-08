cookbook_file "/#{cookbook_name.to_s.tr('-', '_')}_install.txt" do
  source 'taco.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/taco.txt' do
  require 'securerandom'
  source 'taco.txt.erb'
  variables(var: SecureRandom.uuid)
end
