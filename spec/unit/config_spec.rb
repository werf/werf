require_relative '../spec_helper'

describe Dapp::Config::Main do
  def dappfile
    @dappfile ||= ''
  end

  def apps
    Dapp::Config::Main.new(dappfile_path: File.join(Dir.getwd, 'Dappfile')) do |config|
      config.instance_eval(dappfile)
    end._apps
  end

  def app
    apps.first
  end

  def expect_special_attribute(obj, attribute)
    builder = "builder #{obj == :chef ? ':chef' : ':shell'}"
    attribute_path = "#{obj}.#{attribute}"
    @dappfile = %(
      #{builder}
      #{attribute_path} 'a', 'b', 'c'
    )
    expect(app.public_send(obj).public_send("_#{attribute}")).to eq %w(a b c)
    @dappfile = %(
      #{builder}
      #{attribute_path} 'a', 'b', 'c'
      #{attribute_path} 'd', 'e'
    )
    expect(app.public_send(obj).public_send("_#{attribute}")).to eq %w(a b c d e)
  end

  it '#builder' do
    @dappfile = 'builder :chef'
    expect(app._builder).to eq :chef
  end

  it '#builder shell already used' do
    @dappfile = %(
      shell.infra_install 'a'
      chef.module 'a'
    )
    expect { apps }.to raise_error RuntimeError, 'Already defined another builder type!'
  end

  it '#builder chef already used' do
    @dappfile = %(
      builder :chef
      shell.infra_install 'a'
    )
    expect { apps }.to raise_error RuntimeError, 'Already defined another builder type!'
  end

  it '#docker from' do
    @dappfile = "docker.from 'sample'"
    expect(app.docker._from).to eq 'sample'
  end

  it '#docker expose' do
    expect_special_attribute(:docker, :expose)
  end

  it '#chef module' do
    expect_special_attribute(:chef, :module)
  end

  it '#shell attributes' do
    expect_special_attribute(:shell, :infra_install)
    expect_special_attribute(:shell, :infra_setup)
    expect_special_attribute(:shell, :app_install)
    expect_special_attribute(:shell, :app_setup)
  end

  local_attributes = [:cwd, :paths, :owner, :group]
  remote_attributes = local_attributes + [:branch, :ssh_key_path]
  dappfile_local_options = local_attributes.map { |attr| "#{attr}: '#{attr}'" }.join(', ')
  dappfile_remote_options = remote_attributes.map { |attr| "#{attr}: '#{attr}'" }.join(', ')
  dappfile_ga_local = "git_artifact.local 'where_to_add', #{dappfile_local_options}"
  dappfile_ga_remote = "git_artifact.remote 'url', 'where_to_add', #{dappfile_remote_options}"

  [:local, :remote].each do |ga|
    it "#git_artifact #{ga}" do
      attributes = binding.local_variable_get(:"#{ga}_attributes")
      @dappfile = binding.local_variable_get(:"dappfile_ga_#{ga}")
      attributes << :where_to_add
      attributes.each { |attr| expect(app.git_artifact.public_send(ga).first.public_send("_#{attr}")).to eq attr.to_s }
    end
  end

  it '#git_artifact local with remote options' do
    @dappfile = "git_artifact.local 'where_to_add', #{dappfile_remote_options}"
    expect { apps }.to raise_error RuntimeError
  end

  it '#git_artifact name from url' do
    @dappfile = "git_artifact.remote 'https://github.com/flant/dapp.git', 'where_to_add', #{dappfile_remote_options}"
    expect(app.git_artifact.remote.first._name).to eq 'dapp'
  end

  it '#app one' do
    expect(apps.count).to eq 1
    @dappfile = "app 'first'"
    expect(apps.count).to eq 1
    @dappfile = %(
      app 'parent' do
        app 'first'
      end
    )
    expect(apps.count).to eq 1
  end

  it '#app some' do
    @dappfile = %(
      app 'first'
      app 'second'
    )
    expect(apps.count).to eq 2
    @dappfile = %(
      app 'parent' do
        app 'subparent' do
          app 'first'
        end
        app 'second'
      end
    )
    expect(apps.count).to eq 2
  end

  it '#app naming', test_construct: true do
    dir_name = File.basename(Dir.getwd)
    @dappfile = %(
      app 'first'
      app 'parent' do
        app 'subparent' do
          app 'second'
        end
        app 'third'
      end
    )
    expected_apps = ['first', 'parent-subparent-second', 'parent-third'].map { |app| "#{dir_name}-#{app}" }
    expect(apps.map(&:_name)).to eq expected_apps
  end

  it '#app naming with name', test_construct: true do
    @dappfile = %(
      name 'basename'

      app 'first'
      app 'parent' do
        app 'subparent' do
          app 'second'
        end
        app 'third'
      end
    )
    expected_apps = %w(first parent-subparent-second parent-third).map { |app| "basename-#{app}" }
    expect(apps.map(&:_name)).to eq expected_apps
  end

  it '#app naming with name inside app' do
    @dappfile = %(
      app 'parent' do
        app 'subparent' do
          name 'basename'
          app 'second'
        end
      end
    )
    expect { apps }.to raise_error NoMethodError
  end

  it '#app inherit' do
    @dappfile = %(
      docker.from :image_1

      app 'first'
      app 'parent' do
        docker.from :image_2

        app 'subparent' do
          docker.from :image_3
        end
        app 'third'
      end
    )
    expect(apps.map { |app| app.docker._from }).to eq [:image_1, :image_3, :image_2]
  end

  it '#app does not inherit' do
    @dappfile = %(
      app 'first'
      docker.from :image_1
    )
    expect { app.docker._from }.to raise_error RuntimeError, "Docker `from` isn't defined!"
  end

  it '#cache_version' do
    @dappfile = %(
      docker.from :image, cache_version: 'cache_key'
    )
    expect(app.docker._from_cache_version).to eq 'cache_key'
  end
end
