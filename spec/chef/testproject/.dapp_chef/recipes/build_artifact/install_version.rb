file '/version.txt' do
  owner 'root'
  group 'root'
  mode '0777'
  content node.read('ref_name')
  action :create
end
