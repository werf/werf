require_relative '../spec_helper'

describe Dapp::Build::Stage do
  after :each do
    expect(@stages).to be_empty
  end

  def last_stage
    Dapp::Build::Stage::DockerInstructions.new(nil)
  end

  context 'stages' do
    before :each do
      @stages = [:from, :before_install, :g_a_archive_dependencies, :g_a_archive, :'install_group/g_a_pre_patch_dependencies',
                 :'install_group/g_a_pre_patch', :'install_group/install', :'install_group/g_a_post_patch_dependencies',
                 :'install_group/g_a_post_patch', :artifact, :before_setup, :'setup_group/g_a_pre_patch_dependencies',
                 :'setup_group/g_a_pre_patch', :'setup_group/setup', :'setup_group/chef_cookbooks',
                 :'setup_group/g_a_post_patch_dependencies', :'setup_group/g_a_post_patch', :g_a_latest_patch, :docker_instructions]
    end

    def first_stage
      stage = last_stage
      stage = stage.prev_stage while stage.prev_stage
      stage
    end

    it 'prev_stage' do
      stage = last_stage
      while stage
        expect(stage.send(:name)).to eq @stages.pop
        stage = stage.prev_stage
      end
    end

    it 'next_stage' do
      stage = first_stage
      while stage
        expect(stage.send(:name)).to eq @stages.shift
        stage = stage.next_stage
      end
    end
  end

  context 'git_artifact_stages' do
    before :each do
      @stages = [:g_a_archive, :'install_group/g_a_pre_patch', :'install_group/g_a_post_patch',
                 :'setup_group/g_a_pre_patch', :'setup_group/g_a_post_patch', :g_a_latest_patch]
    end

    def last_g_a_stage
      last_stage.prev_stage
    end

    def g_a_archive_stage
      stage = last_g_a_stage
      stage = stage.prev_g_a_stage while stage.send(:name) != :g_a_archive
      stage
    end

    it 'prev_g_a_stage' do
      stage = last_g_a_stage
      while stage.prev_g_a_stage
        expect(stage.send(:name)).to eq @stages.pop
        stage = stage.prev_g_a_stage
      end
      @stages.pop
    end

    it 'next_g_a_stage' do
      stage = g_a_archive_stage
      while stage
        expect(stage.send(:name)).to eq @stages.shift
        stage = stage.next_g_a_stage
      end
    end
  end
end
