directory '/myartifact' do
  owner 'www-data'
  group 'www-data'
  mode '0755'
  action :create
end

file "/myartifact/#{node.read('mdapp-testartifact', 'target_filename')}" do
  owner 'www-data'
  group 'www-data'
  mode '0777'
  content ::File.open("/testartifact/#{node.read('mdapp-testartifact', 'target_filename')}").read
  action :create
end
