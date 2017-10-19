require_relative '../../../spec_helper'

describe Dapp::Dimg::Config::Directive::GitArtifactRemote do
  include SpecHelper::Common
  include SpecHelper::Config::Dimg

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
          expect(dimg_config._git_artifact.public_send("_#{type}").first.public_send("_#{attr}")).to eq(if [:include_paths, :exclude_paths].include? attr
                                                                                                   [attr.to_s]
                                                                                                 else
                                                                                                   attr.to_s
                                                                                                 end)
        end
      end

      context 'stage_dependencies' do
        stage_dependencies_stages = [:install, :setup, :before_setup, :build_artifact]

        it type do
          stages_paths = stage_dependencies_stages.map { |s| [s, %w(a b c).sample] }.to_h
          to_h_expectation = stages_paths.map { |s, p| [s, Array(p)] }.to_h

          dappfile_dimg_git_artifact(type == :local ? :local : 'https://url') do
            add '/cwd' do
              stage_dependencies do
                stage_dependencies_stages.each do |stage|
                  send(stage, stages_paths[stage])
                end
              end
            end
          end

          expect(dimg_config._git_artifact.send("_#{type}").first.send(:stage_dependencies).to_h).to eq(to_h_expectation)
        end

        stage_dependencies_stages.each do |stage|
          it "#{stage} (#{type})" do
            stage_paths = %w(a b)
            to_h_expectation = stage_dependencies_stages.map { |s| [s, []] }.to_h.tap do |h|
              h[stage] = Array(stage_paths)
            end

            dappfile_dimg_git_artifact(type == :local ? :local : 'https://url') do
              add '/cwd' do
                stage_dependencies do
                  send(stage, stage_paths)
                end
              end
            end

            expect(dimg_config._git_artifact.send("_#{type}").first.send(:stage_dependencies).public_send("_#{stage}")).to eq(stage_paths)
            expect(dimg_config._git_artifact.send("_#{type}").first.send(:stage_dependencies).to_h).to eq(to_h_expectation)
          end

          it "#{stage} (#{type}) multiple" do
            dappfile_dimg_git_artifact(type == :local ? :local : 'https://url') do
              add '/cwd' do
                stage_dependencies do
                  send(stage, 'a')
                  send(stage, 'b')
                end
              end
            end

            expect(dimg_config._git_artifact.send("_#{type}").first.send(:stage_dependencies).public_send("_#{stage}")).to eq(%w(a b))
          end
        end
      end
    end

    it 'remote name from url' do
      dappfile_dimg_git_artifact('https://github.com/flant/dapp.git') do
        add '/cwd'
      end
      expect(dimg_config._git_artifact._remote.first._name).to eq 'dapp'
    end

    it 'cwd, to absolute path required' do
      dappfile_dimg_git_artifact(:local) do
        add '/cwd' do
          to '/to'
        end
      end
      expect { dimgs_configs }.to_not raise_error
    end

    it 'include_paths, exclude_paths relative path required' do
      dappfile_dimg_git_artifact(:local) do
        add '/cwd' do
          include_paths 'path1', 'path2'
          exclude_paths 'path1', 'path2'
        end
      end
      expect { dimgs_configs }.to_not raise_error
    end

    it 'default cwd' do
      dappfile_dimg_git_artifact(:local) do
        add
      end
      expect(dimg_config._git_artifact._local.first._cwd).to eq '/'
    end
  end

  context 'negative' do
    it 'cwd absolute path required' do
      dappfile_dimg_git_artifact(:local) do
        add 'cwd'
      end
      expect_exception_code(:export_cwd_absolute_path_required) { dimg_config }
    end

    it 'to absolute path required (2)' do
      dappfile_dimg_git_artifact(:local) do
        add '/cwd' do
          to 'to'
        end
      end
      expect_exception_code(:export_to_absolute_path_required) { dimg_config }
    end

    [:exclude_paths, :include_paths].each do |attr|
      it "#{attr} relative path required (1)" do
        dappfile_dimg_git_artifact(:local) do
          add '/cwd' do
            send(attr, '/path1')
          end
        end
        expect_exception_code(:"export_#{attr}_relative_path_required") { dimg_config }
      end

      it "#{attr} relative path required (2)" do
        dappfile_dimg_git_artifact(:local) do
          add '/cwd' do
            send(attr, 'path1', '/path2')
          end
        end
        expect_exception_code(:"export_#{attr}_relative_path_required") { dimg_config }
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
      expect { dimgs_configs }.to raise_error NoMethodError
    end

    [:local, :remote].each do |type|
      it "stage_dependencies unsupported stage (#{type})" do
        dappfile_dimg_git_artifact(type == :local ? :local : 'https://url') do
          add '/cwd' do
            stage_dependencies do
              unsupported_stage 'a'
            end
          end
        end
        expect { dimgs_configs }.to raise_error NoMethodError
      end

      it "stage_dependencies dependencies must be relative (#{type})" do
        dappfile_dimg_git_artifact(type == :local ? :local : 'https://url') do
          add '/cwd' do
            stage_dependencies do
              install '/a'
            end
          end
        end

        expect_exception_code(:stages_dependencies_paths_relative_path_required) { dimg_config }
      end
    end
  end
end
