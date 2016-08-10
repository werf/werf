log SecureRandom.uuid do
  message "node['test']['common_attr'] = #{node['test']['common_attr']}"
end

cookbook_file "/#{cookbook_name.to_s.tr('-', '_')}_infra_setup.txt" do
  source 'baz.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/baz.txt' do
  require 'securerandom'
  source 'baz.txt.erb'
  variables(var: SecureRandom.uuid)
end
