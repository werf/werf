require 'bundler/gem_tasks'
require 'rspec/core/rake_task'

require 'codeclimate-test-reporter'
CodeClimate::TestReporter.start

RSpec::Core::RakeTask.new(:spec)

task default: :spec
