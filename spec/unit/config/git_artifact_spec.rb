require_relative '../../spec_helper'

describe Dapp::Config::Directive::GitArtifactRemote do
  include SpecHelper::Common
  include SpecHelper::Config

  def dappfile_dimg_git_artifact(type_or_git_repo, &blk)
    dappfile do
      dimg do
        git type_or_git_repo == :local ? nil : type_or_git_repo do
          instance_eval(&blk) if block_given?
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
        dappfile_dimg_git_artifact(type == :local ? :local : 'https://url') do
          add '/cwd' do
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
      dappfile_dimg_git_artifact('https://github.com/flant/dapp.git') do
        add '/cwd'
      end
      expect(dimg._git_artifact._remote.first._name).to eq 'dapp'
    end

    it 'cwd, to absolute path required' do
      dappfile_dimg_git_artifact(:local) do
        add '/cwd' do
          to '/to'
        end
      end
      expect { dimgs }.to_not raise_error
    end

    it 'include_paths, exclude_paths relative path required' do
      dappfile_dimg_git_artifact(:local) do
        add '/cwd' do
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
        add 'cwd'
      end
      expect_exception_code(:export_cwd_absolute_path_required) { dimg }
    end

    it 'to absolute path required (2)' do
      dappfile_dimg_git_artifact(:local) do
        add '/cwd' do
          to 'to'
        end
      end
      expect_exception_code(:export_to_absolute_path_required) { dimg }
    end

    [:exclude_paths, :include_paths].each do |attr|
      it "#{attr} relative path required (1)" do
        dappfile_dimg_git_artifact(:local) do
          add '/cwd' do
            send(attr, '/path1')
          end
        end
        expect_exception_code(:"export_#{attr}_relative_path_required") { dimg }
      end

      it "#{attr} relative path required (2)" do
        dappfile_dimg_git_artifact(:local) do
          add '/cwd' do
            send(attr, 'path1', '/path2')
          end
        end
        expect_exception_code(:"export_#{attr}_relative_path_required") { dimg }
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
