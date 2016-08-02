require_relative '../spec_helper'

describe Dapp::Config::Main do
  include SpecHelper::Expect

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

  def expect_special_attribute(obj, attribute, config_attribute = "_#{attribute}")
    builder = "builder #{obj == :chef ? ':chef' : ':shell'}"
    attribute_setter = "#{obj}.#{attribute}"
    @dappfile = %(
      #{builder}
      #{attribute_setter} 'a', 'b', 'c'
    )
    expect(app.public_send(obj).public_send(config_attribute)).to eq %w(a b c)
    @dappfile = %(
      #{builder}
      #{attribute_setter} 'a', 'b', 'c'
      #{attribute_setter} 'd', 'e'
    )
    expect(app.public_send(obj).public_send(config_attribute)).to eq %w(a b c d e)
  end

  context 'builder' do
    it 'base' do
      @dappfile = 'builder :chef'
      expect(app._builder).to eq :chef
    end

    it 'shell already used (:builder_type_conflict)' do
      @dappfile = %(
        shell.infra_install 'a'
        chef.module 'a'
      )
      expect_exception_code(code: :builder_type_conflict) { apps }
    end

    it 'chef already used (:builder_type_conflict)' do
      @dappfile = %(
        builder :chef
        shell.infra_install 'a'
      )
      expect_exception_code(code: :builder_type_conflict) { apps }
    end
  end

  context 'chef' do
    it 'from' do
      @dappfile = "docker.from 'sample'"
      expect(app.docker._from).to eq 'sample'
    end

    it 'volume' do
      expect_special_attribute(:docker, :volume)
    end

    it 'expose' do
      expect_special_attribute(:docker, :expose)
    end

    it 'env' do
      @dappfile = %(docker.env a: 'b', b: 'c')
      expect(app.docker._env).to eq %w(A=b B=c)
    end

    it 'label' do
      @dappfile = %(docker.label a: 'b', b: 'c')
      expect(app.docker._label).to eq %w(A=b B=c)
    end

    it 'cmd' do
      expect_special_attribute(:docker, :cmd)
    end

    it 'onbuild' do
      expect_special_attribute(:docker, :onbuild)
    end

    it 'workdir' do
      @dappfile = %(
        docker.workdir 'first_value'
        docker.workdir 'second_value'
      )
      expect(app.docker._workdir).to eq 'second_value'
    end

    it 'user' do
      @dappfile = %(
        docker.user 'root'
        docker.user 'root:root'
      )
      expect(app.docker._user).to eq 'root:root'
    end

    it 'module' do
      expect_special_attribute(:chef, :module, :_modules)
    end

    it 'skip_module' do
      expect_special_attribute(:chef, :skip_module, :_skip_modules)
    end

    it 'reset_module' do
      expect_special_attribute(:chef, :reset_module, :_reset_modules)
    end
  end

  context 'shell' do
    def expect_reset_attribute(obj, attribute, config_attribute = "_#{attribute}")
      builder = "builder #{obj == :chef ? ':chef' : ':shell'}"
      attribute_setter = "#{obj}.#{attribute}"
      reset_attribute = "#{obj}.reset_#{attribute}"
      @dappfile = %(
        #{builder}
        #{attribute_setter} 'a', 'b', 'c'
        #{reset_attribute}
      )
      expect(app.public_send(obj).public_send(config_attribute)).to be_empty
    end

    it 'attributes' do
      expect_special_attribute(:shell, :infra_install)
      expect_special_attribute(:shell, :infra_setup)
      expect_special_attribute(:shell, :app_install)
      expect_special_attribute(:shell, :app_setup)
    end

    it 'reset attributes' do
      expect_reset_attribute(:shell, :infra_install)
      expect_reset_attribute(:shell, :infra_setup)
      expect_reset_attribute(:shell, :app_install)
      expect_reset_attribute(:shell, :app_setup)
    end

    it 'reset all attributes' do
      @dappfile = %(
        shell.infra_install 'a'
        shell.infra_setup 'b'
        shell.app_install 'c'
        shell.app_setup 'd'
        shell.reset_all
      )
      [:infra_install, :infra_setup, :app_install, :app_setup].each { |s| expect(app.shell.public_send("_#{s}")).to be_empty }
    end
  end

  context 'git_artifact' do
    local_attributes = [:cwd, :paths, :owner, :group]
    remote_attributes = local_attributes + [:branch, :ssh_key_path]
    dappfile_local_options = local_attributes.map { |attr| "#{attr}: '#{attr}'" }.join(', ')
    dappfile_remote_options = remote_attributes.map { |attr| "#{attr}: '#{attr}'" }.join(', ')
    dappfile_ga_local = "git_artifact.local 'where_to_add', #{dappfile_local_options}"
    dappfile_ga_remote = "git_artifact.remote 'url', 'where_to_add', #{dappfile_remote_options}"

    it 'local' do
      @dappfile = dappfile_ga_local
      local_attributes.delete(:paths)
      expect(app.git_artifact.local.first._paths).to eq ['paths']
      local_attributes.each { |attr| expect(app.git_artifact.local.first.public_send("_#{attr}")).to eq attr.to_s }
    end

    it 'remote' do
      @dappfile = dappfile_ga_remote
      remote_attributes.delete(:paths)
      expect(app.git_artifact.remote.first._paths).to eq ['paths']
      local_attributes.each { |attr| expect(app.git_artifact.remote.first.public_send("_#{attr}")).to eq attr.to_s }
    end

    it 'local with remote options' do
      @dappfile = "git_artifact.local 'where_to_add', #{dappfile_remote_options}"
      expect { apps }.to raise_error Dapp::Error::Config
    end

    it 'git_artifact paths' do
      @dappfile = %( git_artifact.local /where_to_add )
    end

    it 'name from url' do
      @dappfile = "git_artifact.remote 'https://github.com/flant/dapp.git', 'where_to_add', #{dappfile_remote_options}"
      expect(app.git_artifact.remote.first._name).to eq 'dapp'
    end
  end

  context 'app' do
    it 'one' do
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

    it 'some' do
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

    it 'naming', test_construct: true do
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

    it 'naming with name', test_construct: true do
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

    it 'naming with name inside app' do
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

    it 'inherit' do
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

    it 'does not inherit (:docker_from_not_defined)' do
      @dappfile = %(
        app 'first'
        docker.from :image_1
      )
      expect_exception_code(code: :docker_from_not_defined) { app.docker._from }
    end
  end

  context 'cache_version' do
    it 'base' do
      @dappfile = %(
        docker.from :image, cache_version: 'cache_key'
      )
      expect(app.docker._from_cache_version).to eq 'cache_key'
    end
  end
end
