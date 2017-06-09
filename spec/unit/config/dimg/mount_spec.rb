require_relative '../../../spec_helper'

describe Dapp::Dimg::Config::Directive::Mount do
  include SpecHelper::Common
  include SpecHelper::Config::Dimg

  def dappfile_dimg_mount(to, &blk)
    dappfile do
      dimg do
        mount to do
          instance_eval(&blk) if block_given?
        end
      end
    end
  end

  context 'base' do
    [:tmp_dir, :build_dir].each do |type|
      it type do
        dappfile_dimg_mount('/to') do
          from type
        end
        expect(dimg_config.public_send("_#{type}_mount").size).to eq 1
      end
    end

    it :custom_dir do
      dappfile_dimg_mount('/to') do
        from_path '/dir'
      end
      expect(dimg_config._custom_dir_mount.size).to eq 1
      expect(dimg_config._custom_dir_mount.first._from).to eq '/dir'
    end
  end

  context 'negative' do
    it 'type required' do
      dappfile_dimg_mount('/to') do
        from '/from'
      end
      expect_exception_code(:mount_from_type_required) { dimg_config }
    end

    it 'from or from_path required' do
      dappfile_dimg_mount('/to')
      expect_exception_code(:mount_from_or_from_path_required) { dimg_config_validate! }
    end

    it 'duplicate mount point' do
      dappfile do
        dimg do
          mount '/to1' do
            from_path '/from'
          end
          mount '/to1' do
            from_path '/from'
          end
        end
      end

      expect_exception_code(:mount_duplicate_to) { dimg_config_validate! }
    end
  end
end
