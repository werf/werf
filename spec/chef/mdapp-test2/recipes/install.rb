cookbook_file "/#{cookbook_name.to_s.tr('-', '_')}_install.txt" do
  source 'koromyslo.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/koromyslo.txt' do
  require 'securerandom'
  source 'koromyslo.txt.erb'
  variables(var: SecureRandom.uuid)
end
