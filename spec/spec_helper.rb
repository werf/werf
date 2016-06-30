require 'bundler/setup'
require 'test_construct/rspec_integration'

if ENV['CODECLIMATE_REPO_TOKEN']
  require 'codeclimate-test-reporter'
  CodeClimate::TestReporter.start
end

Bundler.require :default, :test, :development

require 'active_support'

require 'spec_helpers/common'
require 'spec_helpers/application'
require 'spec_helpers/git'
require 'spec_helpers/git_artifact'
