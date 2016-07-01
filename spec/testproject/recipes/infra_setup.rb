apt_package 'tree'

cookbook_file '/infra_setup.txt' do
  source 'infra_setup/baz.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end
