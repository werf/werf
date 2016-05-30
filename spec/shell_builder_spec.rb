require_relative 'spec_helper'

describe Dapp::Builder::Shell do
  def shell_builder(**conf)
    conf = { from: 'ubuntu:16.04' }.merge!(conf)
    opts = { docker: Dapp::Docker.new, conf: conf, opts: {} }
    Dapp::Builder::Shell.new(**opts)
  end

  def docker_exec(image, cmd)
    shellout("docker run --rm #{image} bash -lec '#{cmd}'").stdout
  end

  it 'cash: base' do
    builder = shell_builder
    builder.run

    cash = [:infra_install, :infra_setup, :app_install, :app_setup].map { |stage| builder.send("#{stage}_image_name") }

    conf = {}
    conf[:infra_install] = 'date +%s > /infra_install'
    builder = shell_builder(conf)
    builder.run

    [:infra_install, :infra_setup, :app_install, :app_setup].each do |stage|
      expect(cash.include?(builder.send("#{stage}_image_name"))).to_not be_truthy
    end
  end

  it 'cash: container' do
    # base
    conf = {}
    [:infra_install, :infra_setup, :app_install, :app_setup].each { |stage| conf[stage] = "date +%s > /#{stage}" }
    builder = shell_builder(conf)
    builder.run

    [:infra_install, :infra_setup, :app_install, :app_setup].each do |stage|
      expect { docker_exec(builder.send("#{stage}_image_name"), "cat /#{stage}") }.to_not raise_error
    end

    # compare
    app_install_image_name = builder.app_install_image_name
    app_install_timestamp = docker_exec(builder.app_install_image_name, 'cat /app_install')
    expect(docker_exec(builder.app_install_image_name, 'cat /app_install')).to eq(app_install_timestamp)

    conf.delete(:infra_setup)
    builder = shell_builder(conf)
    builder.run

    # infra_setup
    expect { docker_exec(builder.infra_setup_image_name, 'cat /infra_setup') }.to raise_error Mixlib::ShellOut::ShellCommandFailed

    # app_install (next stage)
    expect(builder.app_install_image_name).to_not eq(app_install_image_name)
    expect { docker_exec(builder.app_install_image_name, 'cat /app_install') }.to_not raise_error
    expect(docker_exec(builder.app_install_image_name, 'cat /app_install')).to_not eq(app_install_timestamp)
  end
end
