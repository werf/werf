package 'tree'

cookbook_file "/#{cookbook_name.to_s.tr('-', '_')}_infra_setup.txt" do
  source 'infra_setup/baz.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/baz.txt' do
  require 'securerandom'
  source 'infra_setup/baz.txt.erb'
  variables(var: SecureRandom.uuid)
end
