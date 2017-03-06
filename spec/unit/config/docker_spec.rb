require_relative '../../spec_helper'

describe Dapp::Dimg::Config::Directive::Docker do
  include SpecHelper::Common
  include SpecHelper::Config

  def dappfile_dimg_docker(&blk)
    dappfile do
      dimg do
        docker do
          instance_eval(&blk) if block_given?
        end
      end
    end
  end

  context 'positive' do
    it 'from' do
      dappfile_dimg_docker do
        from 'sample:tag'
      end

      expect(dimg._docker._from).to eq 'sample:tag'
    end

    [:volume, :expose, :cmd, :onbuild].each do |attr|
      it attr do
        expect_array_attribute(attr, method(:dappfile_dimg_docker)) do |*args|
          expect(dimg._docker.send("_#{attr}")).to eq args
        end
      end
    end

    [:env, :label].each do |attr|
      it attr do
        dappfile_dimg_docker do
          send(attr, v1: 1)
        end

        expect(dimg._docker.send("_#{attr}")).to eq(v1: 1)

        dappfile_dimg_docker do
          send(attr, v3: 1)
          send(attr, v2: 1, v1: 1)
        end

        expect(dimg._docker.send("_#{attr}")).to eq(v1: 1, v2: 1, v3: 1)
      end
    end

    [:workdir, :user].each do |attr|
      it attr do
        dappfile_dimg_docker do
          send(attr, 'value1')
          send(attr, 'value2')
        end

        expect(dimg._docker.send("_#{attr}")).to eq 'value2'
      end
    end
  end

  context 'negative' do
    it 'from with incorrect image (:docker_from_incorrect)' do
      dappfile_dimg_docker do
        from "docker.from 'sample'"
      end

      expect_exception_code(:docker_from_incorrect) { dimgs }
    end

    [:env, :label].each do |attr|
      it attr do
        dappfile_dimg_docker do
          send(attr, 'value')
        end

        expect { dimgs }.to raise_error ArgumentError
      end
    end

    [:workdir, :user].each do |attr|
      it attr do
        dappfile_dimg_docker do
          send(attr, 'value1', 'value2')
        end

        expect { dimgs }.to raise_error ArgumentError
      end
    end
  end
end
