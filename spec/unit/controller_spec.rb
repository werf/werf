require_relative '../spec_helper'

describe Dapp::Controller do
  include SpecHelpers::Common
  include SpecHelpers::Expect

  before :each do
    FileUtils.mkdir_p('.dapps/project/config/en')
    FileUtils.touch('.dapps/project/Dappfile')
  end

  RSpec.configure do |c|
    c.before(:example, :build) { stub_application(:build!) }
    c.before(:example, :push) { stub_application(:export!) }
  end

  def stub_application(method)
    stub_instance(Dapp::Application) do |instance|
      allow(instance).to receive(method)
    end
  end

  def controller(cli_options: {}, patterns: nil)
    Dapp::Controller.new(cli_options: { log_color: 'auto' }.merge(cli_options), patterns: patterns)
  end

  it 'build', :build, test_construct: true do
    Pathname('.dapps/project/Dappfile').write("docker.from 'ubuntu.16.04'")
    expect { controller.build }.to_not raise_error
  end

  it 'build:docker_from_not_defined', test_construct: true do
    expect_exception_code(code: :docker_from_not_defined) { controller.build }
  end

  it 'push', :push, test_construct: true do
    expect { controller.push('name') }.to_not raise_error
  end

  it 'push:push_command_unexpected_apps', :push, test_construct: true do
    FileUtils.mkdir_p('.dapps/project2/config/en')
    FileUtils.touch('.dapps/project2/Dappfile')
    expect_exception_code(code: :push_command_unexpected_apps) { controller.push('name') }
  end

  it 'smartpush', :push, test_construct: true do
    expect { controller.smartpush('name') }.to_not raise_error
  end

  it 'list', test_construct: true do
    expect { controller.list }.to_not raise_error
  end

  it 'build_confs (root)', test_construct: true do
    expect { controller(cli_options: { dir: '.dapps/project/' }) }.to_not raise_error
  end

  it 'build_confs (.dapps)', test_construct: true do
    expect { controller }.to_not raise_error
  end

  it 'build_confs (search up)', test_construct: true do
    expect { controller(cli_options: { dir: '.dapps/project/config/en' }) }.to_not raise_error
  end

  it 'build_confs:dappfile_not_found', test_construct: true do
    expect_exception_code(code: :dappfile_not_found) { controller(cli_options: { dir: '.dapps' }) }
  end

  it 'build_confs:no_such_app', test_construct: true do
    expect_exception_code(code: :no_such_app) { controller(patterns: ['app*']) }
  end

  it 'paint_initialize expected cli_options[:log_color] (RuntimeError)' do
    expect { Dapp::Controller.new }.to raise_error RuntimeError
  end
end
