cookbook_file "/#{cookbook_name.to_s.tr('-', '_')}_install.txt" do
  source 'bar.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/bar.txt' do
  require 'securerandom'
  source 'bar.txt.erb'
  variables(var: SecureRandom.uuid)
end
