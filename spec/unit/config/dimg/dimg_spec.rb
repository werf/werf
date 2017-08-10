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
            export '/cwd' do
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

      context 'positive' do
        it 'different to' do
          dappfile_dimg_group_artifact do
            export '/cwd' do
              before :setup
              to '/folder/to1'
            end

            export '/cwd' do
              before :setup
              to '/folder/to2'
            end
          end
          expect { dimg_config_validate! }.to_not raise_error
        end

        it 'different paths' do
          dappfile_dimg_group_artifact do
            export '/cwd' do
              before :setup
              include_paths 'c'
              to '/folder/to1'
            end

            export '/cwd' do
              before :setup
              include_paths 'd'
              to '/folder/to1'
            end
          end
          expect { dimg_config_validate! }.to_not raise_error
        end

        it 'paths with same exclude_paths' do
          dappfile_dimg_group_artifact do
            export '/cwd' do
              before :setup
              include_paths 'c'
              to '/folder/to1'
            end

            export '/cwd' do
              before :setup
              exclude_paths 'c'
              to '/folder/to1'
            end
          end
          expect { dimg_config_validate! }.to_not raise_error
        end

        it 'paths with exclude_paths' do
          dappfile_dimg_group_artifact do
            export '/cwd' do
              before :setup
              include_paths 'c/d/e'
              to '/folder/to1'
            end

            export '/cwd' do
              before :setup
              include_paths 'c'
              exclude_paths 'c/d'
              to '/folder/to1'
            end
          end
          expect { dimg_config_validate! }.to_not raise_error
        end

        it 'to with paths' do
          dappfile_dimg_group_artifact do
            export '/cwd' do
              before :setup
              include_paths 'c'
              to '/folder/to'
            end

            export '/cwd' do
              before :setup
              to '/folder/to/path'
            end
          end
          expect { dimg_config_validate! }.to_not raise_error
        end

        it 'to with exclude_paths' do
          dappfile_dimg_group_artifact do
            export '/cwd' do
              before :setup
              exclude_paths 'path'
              to '/folder/to'
            end

            export '/cwd' do
              before :setup
              to '/folder/to/path'
            end
          end
          expect { dimg_config_validate! }.to_not raise_error
        end
      end

      context 'negative' do
        it 'same to' do
          dappfile_dimg_group_artifact do
            export '/cwd' do
              before :setup
              to '/folder/to'
            end

            export '/cwd' do
              before :setup
              to '/folder/to'
            end
          end
          expect_exception_code(:artifact_conflict) { dimg_config_validate! }
        end

        it 'conflict between to' do
          dappfile_dimg_group_artifact do
            export '/cwd' do
              before :setup
              to '/folder/to'
            end

            export '/cwd' do
              before :setup
              to '/folder/to/path'
            end
          end
          expect_exception_code(:artifact_conflict) { dimg_config_validate! }
        end

        it 'conflict between paths and exclude_paths' do
          dappfile_dimg_group_artifact do
            export '/cwd' do
              before :setup
              include_paths 'c'
              to '/folder/to'
            end

            export '/cwd' do
              before :setup
              exclude_paths 'd'
              to '/folder/to'
            end
          end
          expect_exception_code(:artifact_conflict) { dimg_config_validate! }
        end
      end
    end
  end
end
