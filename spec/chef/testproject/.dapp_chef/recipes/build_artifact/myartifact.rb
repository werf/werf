directory '/myartifact_testproject' do
  owner 'root'
  group 'root'
  mode '0755'
  action :create
end

file "/myartifact_testproject/#{node.read('dimod-testartifact', 'target_filename')}" do
  owner 'root'
  group 'root'
  mode '0777'
  content ::File.open("/testartifact/#{node.read('dimod-testartifact', 'target_filename')}").read
  action :create
end
