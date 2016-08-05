package 'automake'

cookbook_file "/#{cookbook_name.to_s.tr('-', '_')}_setup.txt" do
  source 'pelmeni.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/pelmeni.txt' do
  require 'securerandom'
  source 'pelmeni.txt.erb'
  variables(var: SecureRandom.uuid)
end
