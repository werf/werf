require 'bundler/setup'
require 'test_construct/rspec_integration'

if ENV['CODECLIMATE_REPO_TOKEN']
  require 'codeclimate-test-reporter'
  CodeClimate::TestReporter.start
end

Bundler.require :default, :test, :development

require 'active_support'
require 'recursive_open_struct'

require 'spec_helper/common'
require 'spec_helper/application'
require 'spec_helper/git'

RSpec.configure do |config|
  config.before :all do
    Dapp::Helper::I18n.initialize
  end
  config.mock_with :rspec do |mocks|
    mocks.allow_message_expectations_on_nil = true
  end
end
