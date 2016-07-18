require_relative '../spec_helper'

describe Dapp::Builder::Shell do
  include SpecHelpers::Common
  include SpecHelpers::Application

  def config
    @config ||= default_config.merge(_builder: :shell, _home_path: '')
  end

  def expect_files
    image_name = stages[:source_5].send(:image_name)
    config[:_shell].keys.each do |stage|
      expect { shellout!("docker run --rm #{image_name} bash -lec 'cat /#{stage}'") }.to_not raise_error
    end
  end

  %w(ubuntu:14.04 centos:7).each do |image|
    it "build #{image}" do
      config[:_docker][:_from] = image
      config[:_shell].keys.each { |stage| config[:_shell][stage] << "date +%s > /#{stage}" }
      application_build!

      expect_files
    end
  end
end
