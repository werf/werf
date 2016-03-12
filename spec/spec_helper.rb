require 'bundler/setup'
Bundler.require :default, :test
require 'test_construct/rspec_integration'
require 'codeclimate-test-reporter'

CodeClimate::TestReporter.start
