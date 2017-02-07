name 'test'
version '0.0.1'
depends 'apt'

if ENV['DAPP_CHEF_COOKBOOKS_VENDORING']
  depends 'dimod-test'
  depends 'dimod-test2'
  depends 'dimod-testartifact'
end
