require_relative '../spec_helper'

describe Dapp::CLI do
  include SpecHelper::Common
  include SpecHelper::Stream

  def cli(*args)
    Dapp::CLI.new.run(args)
  rescue SystemExit => _e
    nil
  end

  RSpec.configure do |c|
    c.before(:example, :stub) do
      allow(class_double(Dapp::Application).as_stubbed_const).to receive(:new) { RecursiveOpenStruct.new }
      allow_any_instance_of(Dapp::Controller).to receive(:build_confs) { [RecursiveOpenStruct.new(_name: 'project')] }
    end
  end

  it 'version', test_construct: true do
    expect { cli('--version') }.to output("dapp: #{Dapp::VERSION}\n").to_stdout_from_any_process
  end

  it 'colorize', :stub, test_construct: true do
    out1 = capture_stdout { cli('build', '--color', 'on') }
    out2 = capture_stdout { cli('build', '--color', 'off') }
    expect(out1).to_not eq out2
  end

  it 'log time', :stub, test_construct: true do
    expect { cli('build') }.to_not output(/^[[:digit:]]{4}.[[:digit:]]{2}.[[:digit:]]{2}/).to_stdout_from_any_process
    expect { cli('build', '--time') }.to output(/^[[:digit:]]{4}.[[:digit:]]{2}.[[:digit:]]{2}/).to_stdout_from_any_process
  end
end
