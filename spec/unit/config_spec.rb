require_relative '../spec_helper'

describe Dapp::Config::Main do
  def dappfile
    @dappfile ||= ''
  end

  def apps
    Dapp::Config::Main.new(dappfile_path: File.join(Dir.getwd, 'Dappfile')) do |conf|
      conf.instance_eval(dappfile)
    end._apps
  end

  def app
    apps.first
  end

  def expect_special_attribute(obj, attribute)
    attribute_path = "#{obj}.#{attribute}"
    @dappfile = %{
      #{attribute_path} 'a', 'b', 'c'
    }
    expect(app.public_send(obj).public_send("_#{attribute}")).to eq %w(a b c)
    @dappfile = %{
      #{attribute_path} 'a', 'b', 'c'
      #{attribute_path} 'd', 'e'
    }
    expect(app.public_send(obj).public_send("_#{attribute}")).to eq %w(a b c d e)
  end


  it '#builder' do
    @dappfile = 'builder :chef'
    expect(app._builder).to eq :chef
  end

  it '#builder shell already used' do
    @dappfile = %{
      shell.infra_install 'a'
      chef.module 'a'
    }
    expect { apps }.to raise_error RuntimeError, "Another builder type 'shell' already used!"
  end

  it '#builder chef already used' do
    @dappfile = %{
      builder :chef
      shell.infra_install 'a'
    }
    expect { apps }.to raise_error RuntimeError, "Another builder type 'chef' already used!"
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

  it '#git_artifact local' do
    @dappfile = "git_artifact.local 'where_to_add', #{local_attributes.map { |attr| "#{attr}: '#{attr.to_s}'" }.join(', ')}"
    local_attributes << :where_to_add
    local_attributes.each { |attr| expect(app.git_artifact.local.first.public_send("_#{attr}")).to eq attr.to_s }
  end

  it '#git_artifact local with remote options' do
    @dappfile = "git_artifact.local 'where_to_add', #{remote_attributes.map { |attr| "#{attr}: '#{attr.to_s}'" }.join(', ')}"
    expect { apps }.to raise_error RuntimeError
  end

  it '#git_artifact remote' do
    @dappfile = "git_artifact.remote 'url', 'where_to_add', #{remote_attributes.map { |attr| "#{attr}: '#{attr.to_s}'" }.join(', ')}"
    remote_attributes << :where_to_add
    remote_attributes.each { |attr| expect(app.git_artifact.remote.first.public_send("_#{attr}")).to eq attr.to_s }
  end

  it '#git_artifact name from url' do
    @dappfile = "git_artifact.remote 'https://github.com/flant/dapp.git', 'where_to_add', #{remote_attributes.map { |attr| "#{attr}: '#{attr.to_s}'" }.join(', ')}"
    expect(app.git_artifact.remote.first._name).to eq 'dapp'
  end


  it '#app one' do
    expect(apps.count).to eq 1
    @dappfile = "app 'first'"
    expect(apps.count).to eq 1
    @dappfile = %{
      app 'parent' do
        app 'first'
      end
    }
    expect(apps.count).to eq 1
  end

  it '#app some' do
    @dappfile = %{
      app 'first'
      app 'second'
    }
    expect(apps.count).to eq 2
    @dappfile = %{
      app 'parent' do
        app 'subparent' do
          app 'first'
        end
        app 'second'
      end
    }
    expect(apps.count).to eq 2
  end

  it '#app naming', test_construct: true do
    dir_name = File.basename(Dir.getwd)
    @dappfile = %{
      app 'first'
      app 'parent' do
        app 'subparent' do
          app 'second'
        end
        app 'third'
      end
    }
    expected_apps = ['first', 'parent-subparent-second', 'parent-third'].map { |app| "#{dir_name}-#{app}" }
    expect(apps.map(&:_name)).to eq expected_apps
  end

  it '#app naming with name', test_construct: true do
    @dappfile = %{
      name 'basename'

      app 'first'
      app 'parent' do
        app 'subparent' do
          app 'second'
        end
        app 'third'
      end
    }
    expected_apps = %w(first parent-subparent-second parent-third).map { |app| "basename-#{app}" }
    expect(apps.map(&:_name)).to eq expected_apps


    @dappfile = %{
      app 'parent' do
        app 'subparent' do
          name 'basename'
          app 'second'
        end
      end
    }
    expect(apps.map(&:_name)).to eq ['basename-second']
  end

  it '#app inherit' do
    @dappfile = %{
      docker.from :image_1

      app 'first'
      app 'parent' do
        docker.from :image_2

        app 'subparent' do
          docker.from :image_3
        end
        app 'third'
      end
    }
    expect(apps.map { |app| app.docker._from }).to eq [:image_1, :image_3, :image_2]
  end

  it '#app does not inherit' do
    @dappfile = %{
      app 'first'
      docker.from :image_1
    }
    expect(app.docker._from).to_not eq :image_1
  end


  it '#cache_version' do
    @dappfile = %{
      cache_version from: 'cache_key1'
      cache_version 'cache_key2'
    }
    expect(app._cache_version).to eq 'cache_key2'
    expect(app._cache_version(:from)).to eq 'cache_key1'
  end
end
