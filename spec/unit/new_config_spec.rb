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

    context 'shell' do
      def dappfile_dimg_shell(&blk)
        dappfile do
          dimg do
            shell do
              instance_eval(&blk)
            end
          end
        end
      end

      [:before_install, :before_setup, :install, :setup].each do |attr|
        it attr do
          dappfile_dimg_shell do
            send(attr) do
              command 'cmd'
            end
          end

          expect(dimg.shell.send("_#{attr}")._command).to eq ['cmd']

          dappfile_dimg_shell do
            send(attr) do
              command 'cmd1'
              command 'cmd2', 'cmd3'
            end
          end

          expect(dimg.shell.send("_#{attr}")._command).to eq %w(cmd1 cmd2 cmd3)
        end

        it "#{attr} version" do
          dappfile_dimg_shell do
            send(attr) do
              version 'version'
            end
          end

          expect(dimg.shell.send("_#{attr}")._version).to eq 'version'
        end
      end
    end

    context 'git_artifact' do
      def dappfile_dimg_git_artifact(type_or_git_repo, &blk)
        dappfile do
          dimg do
            git_artifact type_or_git_repo do
              instance_eval(&blk)
            end
          end
        end
      end

      local_artifact_attributes = [:cwd, :include_paths, :exclude_paths, :owner, :group]
      remote_artifact_attributes = local_artifact_attributes + [:branch, :commit]

      context 'positive' do
        [:local, :remote].each do |type|
          it type do
            attributes = binding.local_variable_get("#{type}_artifact_attributes")
            dappfile_dimg_git_artifact(type == :local ? :local : 'url') do
              export '/cwd' do
                attributes.each do |attr|
                  next if attr == :cwd
                  send(attr, attr.to_s)
                end
              end
            end

            attributes.each do |attr|
              next if attr == :cwd
              expect(dimg._git_artifact.public_send("_#{type}").first.public_send("_#{attr}")).to eq(if [:include_paths, :exclude_paths].include? attr
                                                                                                       [attr.to_s]
                                                                                                     else
                                                                                                       attr.to_s
                                                                                                     end)
            end
          end
        end

        it 'remote name from url' do
          dappfile_dimg_git_artifact('https://github.com/flant/dapp.git')
          expect(dimg._git_artifact._remote.first._url).to eq 'dapp'
        end

        it 'cwd, to absolute path required' do
          dappfile_dimg_git_artifact(:local) do
            export '/cwd' do
              to '/to'
            end
          end
          expect { dimgs }.to_not raise_error
        end

        it 'include_paths, exclude_paths relative path required' do
          dappfile_dimg_git_artifact(:local) do
            export '/cwd' do
              include_paths 'path1', 'path2'
              exclude_paths 'path1', 'path2'
            end
          end
          expect { dimgs }.to_not raise_error
        end
      end

      context 'negative' do
        it 'cwd absolute path required' do
          dappfile_dimg_git_artifact(:local) do
            export 'cwd'
          end
          expect { dimgs }.to raise_error
        end

        it 'to absolute path required' do
          dappfile_dimg_git_artifact(:local) do
            export '/cwd' do
              to 'to'
            end
          end
          expect { dimgs }.to raise_error
        end

        [:exclude_paths, :include_paths].each do |attr|
          it "#{attr} relative path required (1)" do
            dappfile_dimg_git_artifact(:local) do
              export '/cwd' do
                send('/path1')
              end
            end
            expect { dimgs }.to raise_error
          end

          it "#{attr} relative path required (2)" do
            dappfile_dimg_git_artifact(:local) do
              export '/cwd' do
                send(attr, 'path1', '/path2')
              end
            end
            expect { dimgs }.to raise_error
          end
        end

        it 'local with remote options' do
          dappfile_dimg_git_artifact(:local) do
            export '/cwd' do
              remote_artifact_attributes.each do |attr|
                next if attr == :cwd
                send(attr, attr.to_s)
              end
            end
          end
          expect { dimgs }.to raise_error NoMethodError
        end
      end
    end
  end
end
