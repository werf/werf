require_relative '../spec_helper'

describe Dapp::Config::DimgGroupMain do
  include SpecHelper::Common
  include SpecHelper::Config

  def project_name
    File.basename(Dir.getwd)
  end

  def dimg_name(name)
    File.join(project_name, name)
  end

  context 'base' do
    context 'positive' do
      it 'dimg without name (1)' do
        dappfile do
          dimg
        end
        expect(dimg._name).to eq project_name
      end

      it 'dimg name' do
        dappfile do
          dimg 'sample'
        end
        expect(dimg._name).to eq dimg_name('sample')
      end
    end

    context 'negative' do
      it 'dimg without name (1)' do
        dappfile do
          dimg
          dimg
        end
        expect { dimgs }.to raise_error RuntimeError
      end

      it 'dimg without name (2)' do
        dappfile do
          dimg_group do
            dimg
            dimg
          end
        end
        expect { dimgs }.to raise_error RuntimeError
      end

      it 'dimg without name (3)' do
        dappfile do
          dimg_group do
            dimg
          end
          dimg_group do
            dimg
          end
        end
        expect { dimgs }.to raise_error RuntimeError
      end
    end
  end

  context 'docker' do
    def dappfile_dimg_docker(&blk)
      dappfile do
        dimg do
          docker do
            instance_eval(&blk)
          end
        end
      end
    end

    context 'positive' do
      it 'from' do
        dappfile_dimg_docker do
          from 'sample:tag'
        end

        expect(dimg.docker._from).to eq 'sample:tag'
      end

      [:volume, :expose, :cmd, :onbuild].each do |attr|
        it attr do
          dappfile_dimg_docker do
            send(attr, 'value')
          end

          expect(dimg.docker.send("_#{attr}")).to eq ['value']

          dappfile_dimg_docker do
            send(attr, 'value3')
            send(attr, 'value1', 'value2')
          end

          expect(dimg.docker.send("_#{attr}")).to eq %w(value3 value1 value2)
        end
      end

      [:env, :label].each do |attr|
        it attr do
          dappfile_dimg_docker do
            send(attr, v1: 1)
          end

          expect(dimg.docker.send("_#{attr}")).to eq ({v1: 1})

          dappfile_dimg_docker do
            send(attr, v3: 1)
            send(attr, v2: 1, v1: 1)
          end

          expect(dimg.docker.send("_#{attr}")).to eq({ v1: 1, v2: 1, v3: 1 })
        end
      end

      [:workdir, :user].each do |attr|
        it attr do
          dappfile_dimg_docker do
            send(attr, 'value1')
            send(attr, 'value2')
          end

          expect(dimg.docker.send("_#{attr}")).to eq 'value2'
        end
      end
    end

    context 'negative' do
      it 'from with incorrect image (:docker_from_incorrect)' do
        dappfile_dimg_docker do
          from "docker.from 'sample'"
        end

        expect_exception_code(code: :docker_from_incorrect) { dimgs }
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
end
