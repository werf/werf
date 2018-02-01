include_recipe 'apt' if node[:platform_family].to_s == 'debian'

package "curl"

file "/myartifact" do
  mode '0777'
  action :create
  content "artifact"
end
