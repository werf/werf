require 'bundler/setup'
require 'test_construct/rspec_integration'

if ENV['CODECLIMATE_REPO_TOKEN']
  require 'codeclimate-test-reporter'
  CodeClimate::TestReporter.start
end

Bundler.require :default, :test, :development

require 'active_support'
require 'recursive_open_struct'

require 'spec_helpers/common'
require 'spec_helpers/application'
require 'spec_helpers/git'
require 'spec_helpers/git_artifact'
require 'spec_helpers/expect'

RSpec.configure do |config|
  config.mock_with :rspec do |mocks|
    mocks.allow_message_expectations_on_nil = true
  end
end
