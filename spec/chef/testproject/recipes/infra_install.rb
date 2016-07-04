apt_package 'cron'

cookbook_file '/infra_install.txt' do
  source 'infra_install/foo.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end

template '/foo.txt' do
  require 'securerandom'
  source 'infra_install/foo.txt.erb'
  variables(foo: "foo:#{SecureRandom.uuid}")
end
