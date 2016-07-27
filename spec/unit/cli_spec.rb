require_relative '../spec_helper'

describe Dapp::CLI do
  include SpecHelpers::Common
  include SpecHelpers::Stream

  def cli(*args)
    Dapp::CLI.new.run(args)
  rescue SystemExit => _e
    nil
  end

  it 'version', test_construct: true do
    expect { cli('--version') }.to output("dapp: #{Dapp::VERSION}\n").to_stdout_from_any_process
  end

  it 'colorize', test_construct: true do
    allow_any_instance_of(Dapp::Application).to receive(:initialize)
    allow_any_instance_of(Dapp::Application).to receive(:build!)
    allow_any_instance_of(Dapp::Controller).to receive(:build_confs) { [RecursiveOpenStruct.new(_name: 'project')] }
    out1 = capture_stdout { cli('build', '--color', 'on') }
    out2 = capture_stdout { cli('build', '--color', 'off') }
    expect(out1).to_not eq out2
  end
end
