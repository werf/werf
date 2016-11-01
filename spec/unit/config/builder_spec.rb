require_relative '../../spec_helper'

describe Dapp::Config::DimgGroupMain do
  include SpecHelper::Common
  include SpecHelper::Config

  context 'positive' do
    it 'base' do
      dappfile do
        dimg_group do
          dimg '1' do
            chef
          end

          dimg '2' do
            shell
          end
        end
      end

      expect(dimg_by_name('1')._builder).to eq :chef
      expect(dimg_by_name('2')._builder).to eq :shell
    end
  end

  context 'negative' do
    it 'builder_type_conflict (1)' do
      dappfile do
        dimg do
          shell
          chef
        end
      end

      expect_exception_code(:builder_type_conflict) { dimg }
    end

    it 'builder_type_conflict (2)' do
      dappfile do
        dimg do
          chef
          shell
        end
      end

      expect_exception_code(:builder_type_conflict) { dimg }
    end

    it 'builder_type_conflict (3)' do
      dappfile do
        dimg_group do
          shell
          chef
        end
      end

      expect_exception_code(:builder_type_conflict) { dimg }
    end

    it 'builder_type_conflict (4)' do
      dappfile do
        dimg_group do
          shell

          dimg 'name' do
            chef
          end
        end
      end

      expect_exception_code(:builder_type_conflict) { dimg }
    end
  end
end
