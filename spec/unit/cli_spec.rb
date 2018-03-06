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
      allow(class_double(Dapp::Dimg::Dimg).as_stubbed_const).to receive(:new) { RecursiveOpenStruct.new }
      allow_any_instance_of(Dapp::Dapp).to receive(:build_configs) { [RecursiveOpenStruct.new(_name: 'dapp')] }
    end
  end

  it 'version' do
    expect { cli('--version') }.to output("dapp: #{Dapp::VERSION}\n").to_stdout_from_any_process
  end

  context 'run' do
    before :each do
      stub_instance(Dapp::Dapp) do |instance|
        allow(instance).to receive(:run)
        @instance = instance
      end
    end

    it 'empty' do
      expect_parsed_options('dimg run')
    end

    it 'dapp args' do
      expect_parsed_options('dimg run --time', options: { time: true })
      expect_parsed_options('dimg run dimg*', options: { dimgs_patterns: ['dimg*'] })
      expect_parsed_options('dimg run dimg* --time', options: { time: true, dimgs_patterns: ['dimg*'] })
      expect_parsed_options('dimg run --time dimg*', options: { time: true, dimgs_patterns: ['dimg*'] })
    end

    it 'docker args' do
      expect_parsed_options('dimg run -ti --rm', docker_options: %w(-ti --rm))
      expect_parsed_options('dimg run -- bash rm -rf', docker_command: %w(bash rm -rf))
      expect_parsed_options('dimg run -ti --rm -- bash rm -rf', docker_options: %w(-ti --rm), docker_command: %w(bash rm -rf))
    end

    it 'oatmeal' do
      expect_parsed_options('dimg run --quiet *dimg* -ti --time --rm -- bash rm -rf',
                            options: { quiet: true, time: true, dimgs_patterns: ['*dimg*'] },
                            docker_options: %w(-ti --rm),
                            docker_command: %w(bash rm -rf))
    end

    def expect_parsed_options(cmd, options: {}, docker_options: [], docker_command: [])
      if docker_options.empty? && docker_command.empty?
        docker_options = %w(-ti --rm)
        docker_command = %w(/bin/bash)
      end

      expect { cli(*cmd.split) }.to_not raise_error
      expect(@instance.options).to include(options)
      expect(@instance.dimgs_patterns).to eq options[:dimgs_patterns] || ['*']
      expect(@instance).to have_received(:run).with(nil, docker_options, docker_command)
    end
  end
end
