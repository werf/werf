require_relative '../spec_helper'

describe Dapp::Artifact do
  include SpecHelper::Common
  include SpecHelper::Application

  def openstruct_config
    @openstruct_config ||= begin
      config[:"_#{@artifact}"].map!(&RecursiveOpenStruct.method(:new))
      RecursiveOpenStruct.new(config)
    end
  end

  def config
    @config ||= default_config.merge(_builder: :shell)
  end

  def artifact_config
    artifact = { _config: Marshal.load(Marshal.dump(config)),
                 _artifact_options: { cwd: '', where_to_add: "/#{@artifact}", exclude_paths: [], paths: [] } }
    artifact[:_config][:_name] = @artifact
    artifact[:_config][:_artifact_dependencies] = []
    artifact[:_config][:_shell][:_build_artifact] = ["mkdir /#{@artifact} && date +%s > /#{@artifact}/test"]
    artifact
  end

  def expect_file
    image_name = stages[expect_stage].send(:image_name)
    expect { shellout!("docker run --rm #{image_name} bash -lec 'cat /#{@artifact}/test'") }.to_not raise_error
  end

  def expect_stage
    (@order == :before) ? @stage : next_stage(@artifact)
  end

  [:before, :after].each do |order|
    [:setup, :install].each do |stage|
      it "build with #{order}_#{stage}_artifact" do
        @artifact = :"#{order}_#{stage}_artifact"
        @order = order
        @stage = stage

        config[:"_#{@artifact}"] = [artifact_config]
        application_build!
        expect_file
      end
    end
  end
end
