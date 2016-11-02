require_relative '../../spec_helper'

describe Dapp::Config::Directive::Mount do
  include SpecHelper::Common
  include SpecHelper::Config

  def dappfile_dimg_mount(type, to)
    dappfile do
      dimg do
        mount to do
          from type
        end
      end
    end
  end

  context 'base' do
    [:tmp_dir, :build_dir].each do |type|
      it type do
        dappfile_dimg_mount(type, '/to')
        expect(dimg.public_send("_#{type}_mount").size).to eq 1
      end
    end

    it 'custom' do
      dappfile_dimg_mount('/from', '/to')
      expect(dimg.public_send('_custom_mount').size).to eq 1
    end
  end

  context 'negative' do
    it 'absolute path required (1)' do
      dappfile_dimg_mount('from', '/to')
      expect_exception_code(:mount_from_absolute_path_required) { dimg }
    end

    it 'absolute path required (2)' do
      dappfile_dimg_mount('/from', 'to')
      expect_exception_code(:mount_to_absolute_path_required) { dimg }
    end
  end
end
