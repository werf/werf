cookbook_file "/#{cookbook_name.to_s.tr('-', '_')}_infra_setup.txt" do
  source 'kolokolchik.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/kolokolchik.txt' do
  require 'securerandom'
  source 'kolokolchik.txt.erb'
  variables(var: SecureRandom.uuid)
end
