template '/X_foo.txt' do
  require 'securerandom'
  source 'foo.txt.erb'
  variables(var: SecureRandom.uuid)
end
