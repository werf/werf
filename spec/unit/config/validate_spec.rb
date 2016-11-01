require_relative '../../spec_helper'

describe Dapp::Config::DimgGroupMain do
  include SpecHelper::Common
  include SpecHelper::Config

  context 'git_artifact' do
    def dappfile_dimg_group_git_artifact(&blk)
      dappfile do
        dimg_group do
          docker do
            from 'image:tag'
          end

          git_artifact :local do
            instance_eval(&blk) if block_given?
          end

          dimg
        end
      end
    end

    it 'to required' do
      dappfile_dimg_group_git_artifact do
        export '/cwd'
      end
      expect_exception_code(:export_to_required) { dimg.send(:validate!) }
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
        expect_exception_code(:export_to_required) { dimg.send(:validate!) }
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
                  to '/to'
                end
              end

              dimg
            end
          end
          expect_exception_code(:scratch_artifact_associated) { dimg.send(:validate!) }
        end
      end

      context 'stage_artifact_not_associated' do
        it 'not associated (:stage_artifact_not_associated)' do
          dappfile do
            dimg_group do
              docker do
                from 'image:tag'
              end

              artifact do
                export '/cwd'
              end

              dimg
            end
          end
          expect_exception_code(:stage_artifact_not_associated) { dimg.send(:validate!) }
        end
      end
    end

    context 'positive' do
      it 'different where_to_add' do
        dappfile_dimg_group_artifact do
          export '/cwd' do
            before :setup
            to '/to1'
          end

          export '/cwd' do
            before :setup
            to '/to2'
          end
        end
        expect { dimg.send(:validate!) }.to_not raise_error
      end

      it 'different paths' do
        dappfile_dimg_group_artifact do
          export '/cwd' do
            before :setup
            include_paths 'c'
            to '/to1'
          end

          export '/cwd' do
            before :setup
            include_paths 'd'
            to '/to1'
          end
        end
        expect { dimg.send(:validate!) }.to_not raise_error
      end

      it 'paths with same exclude_paths' do
        dappfile_dimg_group_artifact do
          export '/cwd' do
            before :setup
            include_paths 'c'
            to '/to1'
          end

          export '/cwd' do
            before :setup
            exclude_paths 'c'
            to '/to1'
          end
        end
        expect { dimg.send(:validate!) }.to_not raise_error
      end

      it 'paths with exclude_paths' do
        dappfile_dimg_group_artifact do
          export '/cwd' do
            before :setup
            include_paths 'c/d/e'
            to '/to1'
          end

          export '/cwd' do
            before :setup
            include_paths 'c'
            exclude_paths 'c/d'
            to '/to1'
          end
        end
        expect { dimg.send(:validate!) }.to_not raise_error
      end

      it 'where_to_add with paths' do
        dappfile_dimg_group_artifact do
          export '/cwd' do
            before :setup
            include_paths 'c'
            to '/to'
          end

          export '/cwd' do
            before :setup
            to '/to/path'
          end
        end
        expect { dimg.send(:validate!) }.to_not raise_error
      end

      it 'where_to_add with exclude_paths' do
        dappfile_dimg_group_artifact do
          export '/cwd' do
            before :setup
            exclude_paths 'path'
            to '/to'
          end

          export '/cwd' do
            before :setup
            to '/to/path'
          end
        end
        expect { dimg.send(:validate!) }.to_not raise_error
      end
    end

    context 'negative' do
      it 'same where_to_add' do
        dappfile_dimg_group_artifact do
          export '/cwd' do
            before :setup
            to '/to'
          end

          export '/cwd' do
            before :setup
            to '/to'
          end
        end
        expect_exception_code(:artifact_conflict) { dimg.send(:validate!) }
      end

      it 'conflict between where_to_add' do
        dappfile_dimg_group_artifact do
          export '/cwd' do
            before :setup
            to '/to'
          end

          export '/cwd' do
            before :setup
            to '/to/path'
          end
        end
        expect_exception_code(:artifact_conflict) { dimg.send(:validate!) }
      end

      it 'conflict between paths and exclude_paths' do
        dappfile_dimg_group_artifact do
          export '/cwd' do
            before :setup
            include_paths 'c'
            to '/to'
          end

          export '/cwd' do
            before :setup
            exclude_paths 'd'
            to '/to'
          end
        end
        expect_exception_code(:artifact_conflict) { dimg.send(:validate!) }
      end
    end
  end
end
