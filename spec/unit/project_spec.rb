require_relative '../spec_helper'

describe Dapp::Project do
  include SpecHelper::Common
  include SpecHelper::Config

  RSpec.configure do |c|
    c.before(:example, :build) { stub_dimg(:build!) }
    c.before(:example, :push) { stub_dimg(:export!) }
  end

  def stub_dimg(method)
    stub_instance(Dapp::Dimg) do |instance|
      allow(instance).to receive(method)
    end
  end

  def stubbed_project(cli_options: {}, dimgs_patterns: nil)
    allow_any_instance_of(Dapp::Project).to receive(:build_configs) {
      [RecursiveOpenStruct.new(_name: 'project'),
       RecursiveOpenStruct.new(_name: 'project2')]
    }
    project(cli_options: cli_options, dimgs_patterns: dimgs_patterns)
  end

  def project(cli_options: {}, dimgs_patterns: nil)
    @project ||= Dapp::Project.new(cli_options: { log_color: 'auto' }.merge(cli_options), dimgs_patterns: dimgs_patterns)
  end

  it 'build', :build, test_construct: true do
    Pathname('Dappfile').write("dimg { docker { from 'ubuntu:16.04' } }")
    expect { project.build }.to_not raise_error
  end

  it 'spush:spush_command_unexpected_dimgs_number', :push do
    expect_exception_code(:spush_command_unexpected_dimgs_number) { stubbed_project.spush('name') }
  end

  it 'run:command_unexpected_dimgs_number', :push do
    expect_exception_code(:command_unexpected_dimgs_number) { stubbed_project.run([], []) }
  end

  it 'list' do
    expect { stubbed_project.list }.to_not raise_error
  end

  it 'paint_initialize expected cli_options[:log_color] (RuntimeError)' do
    expect { Dapp::Project.new }.to raise_error RuntimeError
  end

  context 'build_confs' do
    before :each do
      FileUtils.mkdir_p('dir1/dir2')
      Pathname('Dappfile').write begin
                                   dappfile do
                                     dimg('name') do
                                       docker do
                                         from 'ubuntu:16.04'
                                       end
                                     end
                                   end
                                 end
      allow_any_instance_of(Dapp::Config::Dimg).to receive(:validate!)
    end

    it '.', test_construct: true do
      expect { project.send(:build_configs) }.to_not raise_error
    end

    it 'search up', test_construct: true do
      expect { project(cli_options: { dir: 'dir1/dir2' }).send(:build_configs) }.to_not raise_error
    end

    it 'dappfile_not_found', test_construct: true do
      expect_exception_code(:dappfile_not_found) { project(cli_options: { dir: '..' }).send(:build_configs) }
    end

    it 'no_such_dimg', test_construct: true do
      expect_exception_code(:no_such_dimg) { project(dimgs_patterns: ['dimg*']).send(:build_configs) }
    end
  end
end
