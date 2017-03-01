require_relative '../spec_helper'

describe Dapp::Dimg::Artifact do
  include SpecHelper::Common
  include SpecHelper::Dimg

  def openstruct_config
    @openstruct_config ||= begin
      config[:"_#{@artifact}"].map!(&RecursiveOpenStruct.method(:new))
      RecursiveOpenStruct.new(config)
    end
  end

  def config
    @config ||= begin
      config = default_config.merge(_builder: :shell)
      config[:_shell][:_build_artifact_command] = []
      config
    end
  end

  def artifact_config
    artifact = { _config: Marshal.load(Marshal.dump(config)),
                 _artifact_options: { cwd: "/#{@artifact}", to: "/#{@artifact}", exclude_paths: [], include_paths: [] } }
    artifact[:_config][:_name] = @artifact
    artifact[:_config][:_artifact_dependencies] = []
    artifact[:_config][:_shell][:_build_artifact_command] = ["mkdir /#{@artifact} && date +%s > /#{@artifact}/test"]
    artifact
  end

  context :dimg do
    def expect_file
      image_name = stages[expect_stage].send(:image_name)
      expect { shellout!("docker run --rm #{image_name} bash -lec 'cat /#{@artifact}/test'") }.to_not raise_error
    end

    def expect_stage
      @order == :before ? @stage : next_stage(@artifact)
    end

    [:before, :after].each do |order|
      [:setup, :install].each do |stage|
        it "build with #{order}_#{stage}_artifact" do
          @artifact = :"#{order}_#{stage}_artifact"
          @order = order
          @stage = stage

          config[:"_#{@artifact}"] = [artifact_config]
          dimg_build!
          expect_file
        end
      end
    end
  end

  context :scratch do
    xit 'build with import_artifact' do
      @artifact = :import_artifact
      config[:_import_artifact] = [artifact_config]
      config[:_docker][:_from] = nil
      dimg_build!

      image_name = stages[:import_artifact].send(:image_name)
      container_name = image_name.sub(':', '.')

      begin
        expect do
          shellout!("docker create --name #{container_name} --volume /#{@artifact} #{image_name} no_such_command")
          shellout!("docker run --rm --volumes-from #{container_name} ubuntu:14.04 bash -lec 'cat /#{@artifact}/test'")
        end.to_not raise_error
      ensure
        shellout("docker rm -f #{container_name}")
      end
    end
  end
end
