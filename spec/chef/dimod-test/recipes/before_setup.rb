cookbook_file "/#{cookbook_name.to_s.tr('-', '_')}_before_setup.txt" do
  source 'burger.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/burger.txt' do
  require 'securerandom'
  source 'burger.txt.erb'
  variables(var: SecureRandom.uuid)
end
