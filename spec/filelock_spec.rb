require_relative 'spec_helper'

describe Dapp::Filelock do
  include Dapp::Filelock

  def build_path(x)
    x
  end

  def exit(_x)
    throw :exit
  end

  before :each do
    @builder = instance_double('Dapp::Builder')
    allow(@builder).to receive(:register_atomizer)
  end

  def filelock(**kwargs)
    already_locked = self.class.filelocks['lockfile']

    super 'lockfile', **kwargs do
      expect(File.exist?('lockfile')).to be_truthy
      expect(self.class.filelocks['lockfile']).to be_truthy

      yield if block_given?
    end

    expect(self.class.filelocks['lockfile']).to eq(already_locked)
  end

  it '#simple', test_construct: true do
    filelock
    expect(self.class.filelocks['lockfile']).to be_falsy
  end

  it '#monitor', test_construct: true do
    filelock do
      filelock do
        filelock
      end
      expect(self.class.filelocks['lockfile']).to be_truthy
    end
    expect(self.class.filelocks['lockfile']).to be_falsy
  end

  it '#timeout', test_construct: true do
    filelock do
      self.class.filelocks['lockfile'] = false
      allow(STDERR).to receive(:puts).with('Already in use!')
      expect { filelock(timeout: 0.01) {} }.to throw_symbol(:exit)
      self.class.filelocks['lockfile'] = true
    end
  end
end
