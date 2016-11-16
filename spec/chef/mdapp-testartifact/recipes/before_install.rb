directory '/testartifact' do
  owner 'root'
  group 'root'
  mode '0755'
  action :create
end

cookbook_file "/testartifact/#{node.read('mdapp-testartifact', 'target_filename')}" do
  source 'note.txt'
  owner 'root'
  group 'root'
  mode '0777'
  action :create
end
