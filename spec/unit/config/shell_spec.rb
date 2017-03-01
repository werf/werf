require_relative '../../spec_helper'

describe Dapp::Dimg::Config::Directive::Shell do
  include SpecHelper::Common
  include SpecHelper::Config

  def dappfile_dimg_shell(&blk)
    dappfile do
      dimg do
        shell do
          instance_eval(&blk) if block_given?
        end
      end
    end
  end

  [:before_install, :before_setup, :install, :setup].each do |attr|
    define_method "dappfile_dimg_shell_#{attr}" do |&blk|
      dappfile_dimg_shell do
        send(attr) do
          instance_eval(&blk) unless blk.nil?
        end
      end
    end

    it attr do
      expect_array_attribute(:run, method("dappfile_dimg_shell_#{attr}")) do |*args|
        expect(dimg._shell.send("_#{attr}_command")).to eq args
      end
    end

    it "#{attr} version" do
      dappfile_dimg_shell do
        send(attr) do
          version 'version'
        end
      end

      expect(dimg._shell.send("_#{attr}_version")).to eq 'version'
    end
  end
end
