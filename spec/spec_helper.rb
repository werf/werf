require 'bundler/setup'
require 'test_construct/rspec_integration'

if ENV['CODECLIMATE_REPO_TOKEN']
  require 'codeclimate-test-reporter'
  CodeClimate::TestReporter.start
end

Bundler.require :default, :test, :development

RSpec.configure do |c|
  c.before :all do
    shellout 'git config -l | grep "user.email" || git config --global user.email "dapp@flant.com"'
    shellout 'git config -l | grep "user.name" || git config --global user.name "Dapp Dapp"'
  end
end

def shellout(*args, **kwargs)
  kwargs.delete :log_verbose
  Mixlib::ShellOut.new(*args, timeout: 20, **kwargs).run_command.tap(&:error!)
end
