require_relative '../../../spec_helper'

describe Dapp::Dimg::Config::Directive::Dimg do
  include SpecHelper::Common
  include SpecHelper::Config::Dimg

  context 'naming' do
    context 'positive' do
      it 'dimg without name (1)' do
        dappfile do
          dimg
        end
        expect(dimg_config._name).to eq nil
      end

      it 'dimg name' do
        dappfile do
          dimg 'sample'
        end
        expect(dimg_config._name).to eq 'sample'
      end
    end

    context 'negative' do
      it 'dimg without name (1)' do
        dappfile do
          dimg
          dimg
        end
        expect_exception_code(:dimg_name_required) { dimg_config_validate! }
      end

      it 'dimg without name (2)' do
        dappfile do
          dimg_group do
            dimg
            dimg
          end
        end
        expect_exception_code(:dimg_name_required) { dimg_config_validate! }
      end

      it 'dimg without name (3)' do
        dappfile do
          dimg_group do
            dimg
          end
          dimg_group do
            dimg
          end
        end
        expect_exception_code(:dimg_name_required) { dimg_config_validate! }
      end

      it 'dimg incorrect name' do
        dappfile do
          dimg 'test;'
        end
        expect_exception_code(:dimg_name_incorrect) { dimg_config.validate! }
      end
    end
  end

  context 'builder' do
    context 'positive' do
      it 'base' do
        dappfile do
          dimg_group do
            dimg '1' do
              chef
            end

            dimg '2' do
              shell
            end
          end
        end

        expect(dimg_config_by_name('1')._builder).to eq :chef
        expect(dimg_config_by_name('2')._builder).to eq :shell
      end
    end

    context 'negative' do
      it 'builder_type_conflict (1)' do
        dappfile do
          dimg do
            shell
            chef
          end
        end

        expect_exception_code(:builder_type_conflict) { dimg_config }
      end

      it 'builder_type_conflict (2)' do
        dappfile do
          dimg do
            chef
            shell
          end
        end

        expect_exception_code(:builder_type_conflict) { dimg_config }
      end

      it 'builder_type_conflict (3)' do
        dappfile do
          dimg_group do
            shell
            chef
          end
        end

        expect_exception_code(:builder_type_conflict) { dimg_config }
      end

      it 'builder_type_conflict (4)' do
        dappfile do
          dimg_group do
            shell

            dimg 'name' do
              chef
            end
          end
        end

        expect_exception_code(:builder_type_conflict) { dimg_config }
      end
    end
  end

  context 'validate' do
    context 'git_artifact' do
      def dappfile_dimg_group_git_artifact(&blk)
        dappfile do
          dimg_group do
            docker do
              from 'image:tag'
            end

            git(nil) do
              instance_eval(&blk) if block_given?
            end

            dimg
          end
        end
      end

      it 'to required' do
        dappfile_dimg_group_git_artifact do
          add '/cwd'
        end

        expect_exception_code(:add_to_required) { dimg_config_validate! }
      end
    end

    context 'artifacts' do
      def dappfile_dimg_group_artifact(&blk)
        dappfile do
          dimg_group do
            docker do
              from 'image:tag'
            end

            artifact do
              instance_eval(&blk) if block_given?
            end

            dimg
          end
        end
      end

      context 'artifact' do
        it 'to required' do
          dappfile_dimg_group_artifact do
            export do
              before :setup
            end
          end
          expect_exception_code(:export_to_required) { dimg_config_validate! }
        end

        context 'scratch' do
          it 'associated (:scratch_artifact_associated)' do
            dappfile do
              dimg_group do
                artifact do
                  docker do
                    from 'image:tag'
                  end

                  export '/cwd' do
                    before :setup
                    to '/folder/to'
                  end
                end

                dimg
              end
            end
            expect_exception_code(:scratch_artifact_associated) { dimg_config_validate! }
          end
        end

        it 'stage_artifact_not_associated' do
          dappfile do
            dimg_group do
              docker do
                from 'image:tag'
              end

              artifact do
                export '/cwd' do
                  to '/folder/to'
                end
              end

              dimg
            end
          end
          expect_exception_code(:stage_artifact_not_associated) { dimg_config_validate! }
        end
      end

      def expect_artifact_exports(exports, should_raise_config_error: false)
        [].tap do |sets|
          sets << exports
          sets << exports.reverse unless exports.one?
        end.each do |set|
          expect_dappfile_dimg_group_artifact_exports(set, should_raise_config_error: should_raise_config_error)
        end
      end

      def expect_dappfile_dimg_group_artifact_exports(exports, should_raise_config_error: false)
        dappfile_dimg_group_artifact_exports(exports)

        if should_raise_config_error
          expect { dimg_config_validate! }.to raise_error Dapp::Error::Config
        else
          expect { dimg_config_validate! }.to_not raise_error
        end
      end

      def dappfile_dimg_group_artifact_exports(exports)
        config = [].tap do |lines|
          exports.each do |export|
            lines << "export '/cwd' do"
            lines << "  to '#{export[:to]}'"
            lines << "  include_paths \"#{Array(export[:include_paths]).join('", "')}\"" if export.key?(:include_paths)
            lines << "  exclude_paths \"#{Array(export[:exclude_paths]).join('", "')}\"" if export.key?(:exclude_paths)
            lines << '  before :setup'
            lines << 'end'
          end
        end.join("\n")

        dappfile do
          dimg_group do
            docker do
              from 'image:tag'
            end

            artifact do
              instance_eval(config)
            end

            dimg
          end
        end
      end

      [nil, 'folder'].each do |to_base|
        context "`to` #{to_base.to_s.empty? ? 'without' : 'with'} sub folder" do
          before :all do
            @to_base = to_base
          end

          def to_path(path = nil)
            File.join(['/', @to_base, 'to', path].compact)
          end

          context 'positive' do
            it 'different `to`' do
              exports = begin
                [
                  {
                    to: to_path('folder')
                  },
                  {
                    to: to_path('folder2')
                  }
                ]
              end
              expect_artifact_exports(exports)
            end

            it 'different `include_paths`' do
              exports = begin
                [
                  {
                    to: to_path,
                    include_paths: 'folder'
                  },
                  {
                    to: to_path,
                    include_paths: 'folder2'
                  }
                ]
              end
              expect_artifact_exports(exports)
            end

            it '`paths` & `exclude_paths`' do
              exports = begin
                [
                  {
                    to: to_path,
                    include_paths: 'folder'
                  },
                  {
                    to: to_path,
                    exclude_paths: 'folder'
                  }
                ]
              end
              expect_artifact_exports(exports)
            end

            it '`to` & `exclude_paths`' do
              exports = begin
                [
                  {
                    to: to_path,
                    exclude_paths: 'folder'
                  },
                  {
                    to: to_path('folder')
                  }
                ]
              end
              expect_artifact_exports(exports)
            end
          end

          context 'negative' do
            it 'conflict between `to`' do
              exports = begin
                [
                  {
                    to: to_path,
                  },
                  {
                    to: to_path
                  }
                ]
              end
              expect_artifact_exports(exports, should_raise_config_error: true)
            end

            it 'conflict between `to` (/) and `include_paths`' do
              exports = begin
                [
                  {
                    to: '/folder2'
                  },
                  {
                    to: '/',
                    include_paths: ['folder2']
                  }
                ]
              end
              expect_artifact_exports(exports, should_raise_config_error: true)
            end
          end

          it 'auto excluding (1)' do
            exports = begin
              [
                {
                  to: to_path,
                },
                {
                  to: to_path('path')
                }
              ]
            end
            expect_artifact_exports(exports)
          end

          it 'auto excluding (2)' do
            exports = begin
              [
                {
                  to: to_path,
                  include_paths: 'folder'
                },
                {
                  to: to_path,
                  exclude_paths: 'folder2'
                }
              ]
            end
            expect_artifact_exports(exports)
          end
        end
      end
    end
  end
end
