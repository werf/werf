directory '/myartifact' do
  owner 'root'
  group 'root'
  mode '0755'
  action :create
end

file "/myartifact/#{node['mdapp-testartifact']['target_filename']}" do
  owner 'root'
  group 'root'
  mode '0777'
  content ::File.open('/testartifact/note.txt').read
  action :create
end
