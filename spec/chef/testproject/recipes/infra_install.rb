apt_package 'curl'

cookbook_file "/#{cookbook_name.to_s.gsub('-', '_')}_infra_install.txt" do
  source 'infra_install/foo.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/foo.txt' do
  require 'securerandom'
  source 'infra_install/foo.txt.erb'
  variables(var: SecureRandom.uuid)
end
