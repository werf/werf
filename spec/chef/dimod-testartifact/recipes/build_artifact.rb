directory '/testartifact' do
  owner 'root'
  group 'root'
  mode '0755'
end.run_action(:create)

cookbook_file "/testartifact/#{node.read('dimod-testartifact', 'target_filename')}" do
  source 'note.txt'
  owner 'root'
  group 'root'
  mode '0777'
end.run_action(:create)
