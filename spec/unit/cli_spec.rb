require_relative '../spec_helper'

describe Dapp::CLI do
  include SpecHelper::Common

  def cli(*args)
    Dapp::CLI.new.run(args)
  rescue SystemExit => _e
    nil
  end

  RSpec.configure do |c|
    c.before(:example, :stub) do
      allow(class_double(Dapp::Dimg).as_stubbed_const).to receive(:new) { RecursiveOpenStruct.new }
      allow_any_instance_of(Dapp::Project).to receive(:build_configs) { [RecursiveOpenStruct.new(_name: 'project')] }
    end
  end

  it 'version' do
    expect { cli('--version') }.to output("dapp: #{Dapp::VERSION}\n").to_stdout_from_any_process
  end

  context 'run' do
    before :each do
      stub_instance(Dapp::Project) do |instance|
        allow(instance).to receive(:run)
        @instance = instance
      end
    end

    it 'empty' do
      expect_parsed_options('run')
    end

    it 'project args' do
      expect_parsed_options('run --time', cli_options: { log_time: true })
      expect_parsed_options('run dimg*', dimgs_patterns: ['dimg*'])
      expect_parsed_options('run dimg* --time', cli_options: { log_time: true }, dimgs_patterns: ['dimg*'])
      expect_parsed_options('run --time dimg*', cli_options: { log_time: true }, dimgs_patterns: ['dimg*'])
    end

    it 'docker args' do
      expect_parsed_options('run -ti --rm', docker_options: %w(-ti --rm))
      expect_parsed_options('run -- bash rm -rf', docker_command: %w(bash rm -rf))
      expect_parsed_options('run -ti --rm -- bash rm -rf', docker_options: %w(-ti --rm), docker_command: %w(bash rm -rf))
    end

    it 'oatmeal' do
      expect_parsed_options('run --quiet *dimg* -ti --time --rm -- bash rm -rf',
                            cli_options: { log_quiet: true, log_time: true },
                            dimgs_patterns: ['*dimg*'],
                            docker_options: %w(-ti --rm),
                            docker_command: %w(bash rm -rf))
    end

    def expect_parsed_options(cmd, cli_options: {}, dimgs_patterns: ['*'], docker_options: [], docker_command: [])
      expect { cli(*cmd.split) }.to_not raise_error
      expect(@instance.instance_variable_get(:'@cli_options')).to include(cli_options)
      expect(@instance.instance_variable_get(:'@dimgs_patterns')).to eq dimgs_patterns
      expect(@instance).to have_received(:run).with(docker_options, docker_command)
    end
  end
end
