template '/X_foo.txt' do
  require 'securerandom'
  source 'infra_install/foo.txt.erb'
  variables(var: SecureRandom.uuid)
end
