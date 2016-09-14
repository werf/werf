name 'test'
version '0.0.1'
depends 'apt'

if ENV['DAPP_CHEF_COOKBOOKS_VENDORING']
  depends 'mdapp-test'
  depends 'mdapp-test2'
end
