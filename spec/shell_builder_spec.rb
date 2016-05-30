require_relative 'spec_helper'

describe Dapp::Builder::Base do
  def shell_builder(**options)
    opts = { docker: Dapp::Docker.new, conf: { from: 'ubuntu:16.04' }, opts: {} }.merge(options)
    Dapp::Builder::Shell.new(**opts)
  end

  def docker_exec(image, cmd)
    shellout("docker run --rm #{image} bash -lec '#{cmd}'")
  end

  it 'cash' do
    builder = shell_builder
    builder.run

    cash = [:infra_install, :infra_setup, :app_install, :app_setup].map { |stage| puts builder.send("#{stage}_image_name") }
  end

  it 'check stage' do
    [:infra_install, :infra_setup, :app_install, :app_setup].each { |stage| options[:conf][stage] = "date +%s > /#{stage}" }
    builder = shell_builder(options)
    builder.run



    # check shell stages
    [:infra_install, :infra_setup, :app_install, :app_setup].each do |stage|
      expect { docker_exec(builder.send("#{stage}_image_name"), "cat /#{stage}") }.to_not raise_error
    end

    app_setup_image_name = builder.app_setup_image_name
    app_setup_timestamp = docker_exec(builder.app_setup_image_name, 'cat /app_setup').stdout

    expect(docker_exec(builder.app_setup_image_name, 'cat /app_setup').stdout).to eq(app_setup_timestamp)

    # check rehash
    options[:conf].delete(:infra_setup)
    builder = Dapp::Builder::Shell.new(**options)
    builder.run

    [:infra_install, :infra_setup, :app_install, :app_setup].each do |stage|
      puts builder.send("#{stage}_image_name")
    end

    expect { docker_exec(builder.app_install_image_name, 'cat /infra_setup') }.to raise_error Mixlib::ShellOut::ShellCommandFailed
    expect(builder.app_setup_image_name).to_not eq(app_setup_image_name)
    expect { docker_exec(builder.app_setup_image_name, 'cat /app_setup').stdout }.to_not raise_error
    expect(docker_exec(builder.app_setup_image_name, 'cat /app_setup').stdout).to_not eq(app_setup_timestamp)
  end
end
