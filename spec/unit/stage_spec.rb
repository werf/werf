require_relative '../spec_helper'

describe Dapp::Build::Stage do
  def last_stage
    Dapp::Build::Stage::Source5.new(nil)
  end

  after :each do
    expect(@stages).to be_empty
  end

  context 'stages' do
    before :each do
      @stages = [:from, :infra_install, :source_1_archive_dependencies, :source_1_archive, :source_1_dependencies,
                 :source_1, :install, :artifact, :source_2_dependencies, :source_2, :infra_setup, :source_3_dependencies, :source_3,
                 :setup, :chef_cookbooks, :source_4_dependencies, :source_4, :source_5]
    end

    def first_stage
      stage = last_stage
      while stage.prev_stage; stage = stage.prev_stage end
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

  context 'source_stages' do
    before :each do
      @stages = [:source_1_archive, :source_1, :source_2, :source_3, :source_4, :source_5]
    end

    def source_1_archive_stage
      stage = last_stage
      while stage.send(:name) != :source_1_archive; stage = stage.prev_source_stage end
      stage
    end

    it 'prev_source_stage' do
      stage = last_stage
      while stage.prev_source_stage
        expect(stage.send(:name)).to eq @stages.pop
        stage = stage.prev_source_stage
      end
      @stages.pop
    end

    it 'next_source_stage' do
      stage = source_1_archive_stage
      while stage
        expect(stage.send(:name)).to eq @stages.shift
        stage = stage.next_source_stage
      end
    end
  end
end
