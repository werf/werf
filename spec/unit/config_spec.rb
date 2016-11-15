require_relative '../spec_helper'

describe Dapp::Config::Main do
  include SpecHelper::Common

  def dappfile
    @dappfile ||= ''
  end

  def apps
    Dapp::Config::Main.new(dappfile_path: File.join(Dir.getwd, 'Dappfile'), project: stubbed_project) do |config|
      config.instance_eval(dappfile)
    end._apps
  end

  def stubbed_project
    instance_double(Dapp::Project).tap do |instance|
      allow(instance).to receive(:log_warning)
    end
  end

  def app
    apps.first
  end

  def apps_by_name
    apps.map { |app| [app._name, app] }.to_h
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
        shell.before_install 'a'
        chef.module 'a'
      )
      expect_exception_code(code: :builder_type_conflict) { apps }
    end

    it 'chef already used (:builder_type_conflict)' do
      @dappfile = %(
        builder :chef
        shell.before_install 'a'
      )
      expect_exception_code(code: :builder_type_conflict) { apps }
    end
  end

  context 'docker' do
    it 'from' do
      @dappfile = "docker.from 'sample:tag'"
      expect(app.docker._from).to eq 'sample:tag'
    end

    it 'from with incorrect image (:docker_from_incorrect)' do
      @dappfile = "docker.from 'sample'"
      expect_exception_code(code: :docker_from_without_tag) { apps }
    end

    it 'volume' do
      expect_special_attribute(:docker, :volume)
    end

    it 'expose' do
      expect_special_attribute(:docker, :expose)
    end

    it 'env' do
      @dappfile = %(docker.env a: 'b', b: 'c')
      expect(app.docker._env).to eq(a: 'b', b: 'c')
    end

    it 'label' do
      @dappfile = %(docker.label a: 'b', b: 'c')
      expect(app.docker._label).to eq(a: 'b', b: 'c')
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
  end

  context :chef do
    it 'module' do
      expect_special_attribute(:chef, :module, :_modules)
    end

    it 'skip_module' do
      @dappfile = %(
        builder :chef

        chef.module 'a', 'b', 'c', 'd'
        chef.skip_module 'a', 'c'

        app 'X' do
          chef.module 'e', 'f'
        end

        app 'Y' do
          chef.module 'g'
          chef.skip_module 'b'
        end
      )

      expect(apps_by_name['dapp-X'].chef._modules).to eq %w(b d e f)
      expect(apps_by_name['dapp-Y'].chef._modules).to eq %w(d g)
    end

    it 'reset_modules' do
      @dappfile = %(
        builder :chef

        chef.module 'a', 'b', 'c'

        app 'X' do
          chef.reset_modules
        end

        app 'Y' do
          chef.module 'd'

          app 'A' do
            chef.reset_modules
          end

          app 'B'
        end

        chef.reset_modules

        app 'Z'
      )

      expect(apps_by_name['dapp-X'].chef._modules).to eq %w()
      expect(apps_by_name['dapp-Y-A'].chef._modules).to eq %w()
      expect(apps_by_name['dapp-Y-B'].chef._modules).to eq %w(a b c d)
      expect(apps_by_name['dapp-Z'].chef._modules).to eq %w()
    end

    it 'recipe' do
      expect_special_attribute(:chef, :recipe, :_recipes)
    end

    it 'remove_recipe' do
      @dappfile = %(
        builder :chef

        chef.recipe 'a', 'b', 'c', 'd'
        chef.remove_recipe 'a', 'c'

        app 'X' do
          chef.recipe 'e', 'f'
        end

        app 'Y' do
          chef.recipe 'g'
          chef.remove_recipe 'b'
        end
      )

      expect(apps_by_name['dapp-X'].chef._recipes).to eq %w(b d e f)
      expect(apps_by_name['dapp-Y'].chef._recipes).to eq %w(d g)
    end

    it 'reset_recipes' do
      @dappfile = %(
        builder :chef

        chef.recipe 'a', 'b', 'c'

        app 'X' do
          chef.reset_recipes
        end

        app 'Y' do
          chef.recipe 'd'

          app 'A' do
            chef.reset_recipes
          end

          app 'B'
        end

        chef.reset_recipes

        app 'Z'
      )

      expect(apps_by_name['dapp-X'].chef._recipes).to eq %w()
      expect(apps_by_name['dapp-Y-A'].chef._recipes).to eq %w()
      expect(apps_by_name['dapp-Y-B'].chef._recipes).to eq %w(a b c d)
      expect(apps_by_name['dapp-Z'].chef._recipes).to eq %w()
    end

    it 'attributes' do
      @dappfile = %(
        builder :chef

        chef.attributes['k1']['k2'] = 'k1k2value'
        chef.attributes['k1']['k3'] = 'k1k3value'
#{%i(before_install install before_setup setup build_artifact).map do |stage|
  "chef.#{stage}_attributes['k1']['#{stage}'] = 'k1#{stage}value'"
end.join("\n")}

        app 'X' do
          chef.attributes['k1'].delete('k2')
          chef.attributes['k1'].delete('k3')
          chef.attributes['k1']['k4']['k5'] = 'k1k4k5value'
          chef.before_install_attributes['k1']['before_install'] = 'k1before_installchangedvalue'
        end

        app 'Y' do
          chef.attributes['k1']['k2'] = 'k1k2changedvalue'
          chef.attributes['k1']['k4'] = 'k1k4value'

          app 'A' do
            chef.attributes['k1']['k5'] = 'k1k5value'
          end
        end

        app 'Z'

        app 'E' do
          chef.reset_attributes
        end

        app 'T' do
          chef.reset_before_install_attributes
        end

        app 'S' do
          chef.reset_all
        end

        app 'Q' do
          chef.reset_all_attributes
        end
      )

      %w(before_install install before_setup setup build_artifact).each do |stage|
        expect(apps_by_name['dapp-X'].chef.send("_#{stage}_attributes")).to eq({
          'k1' => {
            'k4' => { 'k5' => 'k1k4k5value' },
            stage => (stage == 'before_install' ? "k1#{stage}changedvalue" : "k1#{stage}value")
          }
        })
        expect(apps_by_name['dapp-Y-A'].chef.send("_#{stage}_attributes")).to eq({
          'k1' => {
            'k2' => 'k1k2changedvalue',
            'k3' => 'k1k3value',
            'k4' => 'k1k4value',
            'k5' => 'k1k5value',
            stage => "k1#{stage}value"
          }
        })
        expect(apps_by_name['dapp-Z'].chef.send("_#{stage}_attributes")).to eq({
          'k1' => {
            'k2' => 'k1k2value',
            'k3' => 'k1k3value',
            stage => "k1#{stage}value"
          }
        })
        expect(apps_by_name['dapp-E'].chef.send("_#{stage}_attributes")).to eq({
          'k1' => {
            stage => "k1#{stage}value"
          }
        })
        expect(apps_by_name['dapp-T'].chef.send("_#{stage}_attributes")).to eq({
          'k1' => {
            'k2' => 'k1k2value',
            'k3' => 'k1k3value'
          }.tap { |obj|
            obj[stage] = "k1#{stage}value" if stage != 'before_install'
          }
        })
        expect(apps_by_name['dapp-S'].chef.send("_#{stage}_attributes")).to eq({})
        expect(apps_by_name['dapp-Q'].chef.send("_#{stage}_attributes")).to eq({})
      end
    end

    it 'reset_all' do # FIXME: attributes
      @dappfile = %(
        builder :chef

        chef.module 'ma', 'mb', 'mc'
        chef.recipe 'ra', 'rb', 'rc'
        chef.attributes['k1']['k2'] = 'k1k2value'

        app 'X' do
          chef.module 'md'
          chef.recipe 'rd'
          chef.attributes['k1']['k3'] = 'k1k3value'

          app 'A'

          app 'B' do
            chef.reset_all
          end

          chef.reset_all

          app 'C'
        end

        app 'Y'

        chef.reset_all

        app 'Z'
      )

      expect(apps_by_name['dapp-X-A'].chef._modules).to eq %w(ma mb mc md)
      expect(apps_by_name['dapp-X-A'].chef._recipes).to eq %w(ra rb rc rd)
      expect(apps_by_name['dapp-X-A'].chef._before_install_attributes).to eq({
        'k1' => {
          'k2' => 'k1k2value',
          'k3' => 'k1k3value'
        }
      })

      expect(apps_by_name['dapp-X-B'].chef._modules).to eq %w()
      expect(apps_by_name['dapp-X-B'].chef._recipes).to eq %w()
      expect(apps_by_name['dapp-X-B'].chef._before_install_attributes).to eq({})

      expect(apps_by_name['dapp-X-C'].chef._modules).to eq %w()
      expect(apps_by_name['dapp-X-C'].chef._recipes).to eq %w()
      expect(apps_by_name['dapp-X-C'].chef._before_install_attributes).to eq({})

      expect(apps_by_name['dapp-Y'].chef._modules).to eq %w(ma mb mc)
      expect(apps_by_name['dapp-Y'].chef._recipes).to eq %w(ra rb rc)
      expect(apps_by_name['dapp-Y'].chef._before_install_attributes).to eq({
        'k1' => {
          'k2' => 'k1k2value'
        }
      })

      expect(apps_by_name['dapp-Z'].chef._modules).to eq %w()
      expect(apps_by_name['dapp-Z'].chef._recipes).to eq %w()
      expect(apps_by_name['dapp-Z'].chef._before_install_attributes).to eq({})
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
      expect_special_attribute(:shell, :before_install)
      expect_special_attribute(:shell, :before_setup)
      expect_special_attribute(:shell, :install)
      expect_special_attribute(:shell, :setup)
    end

    it 'reset attributes' do
      expect_reset_attribute(:shell, :before_install)
      expect_reset_attribute(:shell, :before_setup)
      expect_reset_attribute(:shell, :install)
      expect_reset_attribute(:shell, :setup)
    end

    it 'reset all attributes' do
      @dappfile = %(
        shell.before_install 'a'
        shell.before_setup 'b'
        shell.install 'c'
        shell.setup 'd'
        shell.reset_all
      )
      [:before_install, :before_setup, :install, :setup].each { |s| expect(app.shell.public_send("_#{s}")).to be_empty }
    end
  end

  context 'artifact' do
    before :each do
      @artifact_attributes = [:cwd, :paths, :exclude_paths, :owner, :group]
    end

    context 'base' do
      it 'unsupported inherited artifact' do
        @dappfile = "artifact('where_to_add') { artifact 'where_to_add' }"
        expect { apps }.to raise_error NameError
      end

      it 'unsupported key (:artifact_unexpected_attribute)' do
        @dappfile = "artifact 'where_to_add', unsupported_key: :value"
        expect_exception_code(code: :artifact_unexpected_attribute) { apps }
      end

      it 'unsupported app method in artifact body' do
        @dappfile = "artifact('where_to_add', unsupported_key: :value) { app }"
        expect { apps }.to raise_error NameError
      end

      [:volume, :expose, :env, :label, :cmd, :onbuild, :workdir, :user, :entrypoint].each do |instruction|
        it "unsupported docker.#{instruction} in artifact body" do
          @dappfile = "artifact('where_to_add', unsupported_key: :value) { docker.#{instruction} }"
          expect { apps }.to raise_error NoMethodError
        end
      end

      it 'artifact_dependencies' do
        @dappfile = %(
          artifact 'where_to_add' do
            artifact_depends_on 'depends'
          end
        )
        expect(app._import_artifact.first._config._artifact_dependencies).to eq ['depends']
      end

      it 'shell.build_artifact' do
        @dappfile = %(
          artifact 'where_to_add' do
            shell.build_artifact 'cmd'
          end
        )
        expect(app._import_artifact.first._config._shell._build_artifact).to eq ['cmd']
      end

      context 'not scratch' do
        before :each do
          @dappfile = "docker.from 'image:tag'\n"
        end

        it 'double associated (:stage_artifact_double_associate)' do
          @dappfile << "artifact 'where_to_add', before: :setup, after: :setup"
          expect_exception_code(code: :stage_artifact_double_associate) { apps }
        end

        it 'not supported associated stage (:stage_artifact_not_supported_associated_stage)' do
          @dappfile << "artifact 'where_to_add', before: :from"
          expect_exception_code(code: :stage_artifact_not_supported_associated_stage) { apps }
        end
      end
    end

    [:before, :after].each do |order|
      [:setup, :install].each do |stage|
        artifact = "#{order}_#{stage}_artifact"
        context artifact do
          it 'base' do
            @dappfile = "artifact 'where_to_add', #{order}: :#{stage}, #{@artifact_attributes.map { |attr| "#{attr}: '#{attr}'" }.join(', ')}"
            [:paths, :exclude_paths].each do |attr|
              @artifact_attributes.delete(attr)
              expect(app.public_send(:"_#{artifact}").first.public_send("_#{attr}")).to eq [attr.to_s]
            end
            expect(app.public_send(:"_#{artifact}").first._paths).to eq ['paths']
            @artifact_attributes.each { |attr| expect(app.public_send(:"_#{artifact}").first.public_send("_#{attr}")).to eq attr.to_s }
          end
        end
      end
    end

    context 'git_artifact' do
      before :each do
        @remote_attributes = @artifact_attributes + [:branch]
        @dappfile_remote_options = @remote_attributes.map { |attr| "#{attr}: '#{attr}'" }.join(', ')
      end

      it 'remote' do
        @dappfile = "git_artifact.remote 'url', 'where_to_add', #{@dappfile_remote_options}"
        [:paths, :exclude_paths].each do |attr|
          @remote_attributes.delete(attr)
          expect(app.git_artifact.remote.first.public_send("_#{attr}")).to eq [attr.to_s]
        end
        @remote_attributes.each { |attr| expect(app.git_artifact.remote.first.public_send("_#{attr}")).to eq attr.to_s }
      end

      it 'local with remote options (:git_artifact_unexpected_attribute)' do
        @dappfile = "git_artifact.local 'where_to_add', #{@dappfile_remote_options}"
        expect_exception_code(code: :git_artifact_unexpected_attribute) { apps }
      end

      it 'git_artifact paths' do
        @dappfile = %( git_artifact.local /where_to_add )
      end

      it 'name from url' do
        @dappfile = "git_artifact.remote 'https://github.com/flant/dapp.git', 'where_to_add', #{@dappfile_remote_options}"
        expect(app.git_artifact.remote.first._name).to eq 'dapp'
      end
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
        docker.from 'image_1:tag'

        app 'first'
        app 'parent' do
          docker.from 'image_2:tag'

          app 'subparent' do
            docker.from 'image_3:tag'
          end
          app 'third'
        end
      )
      expect(apps.map { |app| app.docker._from }).to eq %w(image_1:tag image_3:tag image_2:tag)
    end

    it 'does not inherit' do
      @dappfile = %(
        app 'first'
        docker.from 'image:tag'
      )
      expect(app.docker._from).to be_nil
    end
  end

  context 'basename' do
    it 'base' do
      expect(app._basename).to eq 'dapp'
    end

    it 'name' do
      @dappfile = "name 'test'"
      expect(app._basename).to eq 'test'
    end

    it 'incorrect name (:app_name_incorrect)' do
      @dappfile = "name 'test;'"
      expect_exception_code(code: :app_name_incorrect) { app._name }
    end
  end

  [:tmp_dir, :build_dir].each do |dir|
    context dir do
      it 'base' do
        @dappfile = "#{dir}.store 'test'"
        expect(app.public_send("_#{dir}")._store).to eq ['test']
      end

      it 'unstore' do
        @dappfile = %(
          #{dir}.store 'test', 'test2'
          #{dir}.unstore 'test'
        )
        expect(app.public_send("_#{dir}")._store).to eq ['test2']
      end

      it 'reset' do
        @dappfile = %(
          #{dir}.store 'test', 'test2'
          #{dir}.reset
        )
        expect(app.public_send("_#{dir}")._store).to be_empty
      end
    end
  end

  context 'cache_version' do
    it 'base' do
      @dappfile = %(
        docker.from 'image:tag', cache_version: 'cache_key'
      )
      expect(app.docker._from_cache_version).to eq 'cache_key'
    end
  end

  context 'validate' do
    context 'artifact' do
      context 'scratch' do
        it 'associated (:scratch_artifact_associated)' do
          @dappfile = "artifact 'where_to_add', before: :setup"
          expect_exception_code(code: :scratch_artifact_associated) { app.send(:validate!) }
        end
      end

      context 'not scratch' do
        it 'not associated (:stage_artifact_not_associated)' do
          @dappfile = %(
            docker.from 'image:tag'
            artifact 'where_to_add'
          )
          expect_exception_code(code: :stage_artifact_not_associated) { app.send(:validate!) }
        end
      end
    end

    context 'artifacts' do
      context 'positive' do
        it 'different where_to_add' do
          @dappfile = %(
            docker.from 'image:tag'

            artifact 'where_to_add1', before: :setup
            artifact 'where_to_add2', after: :install
          )
          expect { app.send(:validate!) }.to_not raise_error
        end

        it 'different paths' do
          @dappfile = %(
            docker.from 'image:tag'

            artifact 'where_to_add', paths: "c", before: :setup
            artifact 'where_to_add', paths: "d", after: :install
          )
          expect { app.send(:validate!) }.to_not raise_error
        end

        it 'paths with same exclude_paths' do
          @dappfile = %(
            docker.from 'image:tag'

            artifact 'where_to_add', paths: "c", before: :setup
            artifact 'where_to_add', exclude_paths: "c", after: :install
          )
          expect { app.send(:validate!) }.to_not raise_error
        end

        it 'paths with exclude_paths' do
          @dappfile = %(
            docker.from 'image:tag'

            artifact 'where_to_add', paths: "c/d/e", before: :setup
            artifact 'where_to_add', paths: "c", exclude_paths: "c/d", after: :install
          )
          expect { app.send(:validate!) }.to_not raise_error
        end

        it 'where_to_add with paths' do
          @dappfile = %(
            docker.from 'image:tag'

            artifact 'where_to_add', paths: "c", before: :setup
            artifact 'where_to_add/ddd', after: :install
          )
          expect { app.send(:validate!) }.to_not raise_error
        end

        it 'where_to_add with exclude_paths' do
          @dappfile = %(
            docker.from 'image:tag'

            artifact 'where_to_add', exclude_paths: "ddd", before: :setup
            artifact 'where_to_add/ddd', after: :install
          )
          expect { app.send(:validate!) }.to_not raise_error
        end
      end

      context 'negative' do
        it 'same where_to_add' do
          @dappfile = %(
            docker.from 'image:tag'

            artifact 'where_to_add', before: :setup
            artifact 'where_to_add', after: :install
          )
          expect_exception_code(code: :artifact_conflict) { app.send(:validate!) }
        end

        it 'conflict between where_to_add' do
          @dappfile = %(
            docker.from 'image:tag'

            artifact 'where_to_add', before: :setup
            artifact 'where_to_add/ddd', after: :install
          )
          expect_exception_code(code: :artifact_conflict) { app.send(:validate!) }
        end

        it 'conflict between paths and exclude_paths' do
          @dappfile = %(
            docker.from 'image:tag'

            artifact 'where_to_add', paths: "c", before: :setup
            artifact 'where_to_add', exclude_paths: "d", after: :install
          )
          expect_exception_code(code: :artifact_conflict) { app.send(:validate!) }
        end
      end
    end
  end
end
