require_relative '../spec_helper'

describe Dapp::Controller do
  include SpecHelpers::Common
  include SpecHelpers::Expect

  before :each do
    FileUtils.mkdir_p('.dapps/project/config/en')
    FileUtils.touch('.dapps/project/Dappfile')
  end

  def stub_application(method)
    stub_instance(Dapp::Application) do |instance|
      allow(instance).to receive(method)
    end
  end

  it 'build', test_construct: true do
    stub_application(:build!)
    Pathname('.dapps/project/Dappfile').write("docker.from 'ubuntu.16.04'")
    expect { Dapp::Controller.new(cli_options: {}).build }.to_not raise_error
  end

  it 'build (:docker_from_not_defined)', test_construct: true do
    expect_exception_code(code: :docker_from_not_defined) { Dapp::Controller.new(cli_options: {}).build }
  end

  it 'push', test_construct: true do
    stub_application(:export!)
    expect { Dapp::Controller.new(cli_options: {}).push('name') }.to_not raise_error
  end

  it 'push (:push_command_unexpected_apps)', test_construct: true do
    FileUtils.mkdir_p('.dapps/project2/config/en')
    FileUtils.touch('.dapps/project2/Dappfile')

    stub_application(:export!)
    expect_exception_code(code: :push_command_unexpected_apps) { Dapp::Controller.new(cli_options: {}).push('name') }
  end

  it 'smartpush', test_construct: true do
    stub_application(:export!)
    expect { Dapp::Controller.new(cli_options: {}).smartpush('name') }.to_not raise_error
  end

  it 'list', test_construct: true do
    expect { Dapp::Controller.new(cli_options: {}).list }.to_not raise_error
  end

  it 'build_confs (root)', test_construct: true do
    expect { Dapp::Controller.new(cli_options: { dir: '.dapps/project/' }) }.to_not raise_error
  end

  it 'build_confs (.dapps)', test_construct: true do
    expect { Dapp::Controller.new(cli_options: {}) }.to_not raise_error
  end

  it 'build_confs (search up)', test_construct: true do
    expect { Dapp::Controller.new(cli_options: { dir: '.dapps/project/config/en' }) }.to_not raise_error
  end
end
