apt_package 'unzip'

cookbook_file '/app_setup.txt' do
  source 'app_setup/qux.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end
