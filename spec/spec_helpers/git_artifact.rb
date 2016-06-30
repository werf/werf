module SpecHelpers
  module GitArtifact
    # rubocop:disable Metrics/ParameterLists
    def git_artifact_init(where_to_add, id: nil, changefile: 'data.txt', changedata: random_string, branch: 'master', **kwargs)
      repo_change_and_commit(changefile: changefile, changedata: changedata, branch: branch)
      (@git_artifacts ||= {})[id] = Dapp::GitArtifact.new(repo, where_to_add, branch: branch, **kwargs)
    end
    # rubocop:enable Metrics/ParameterLists

    def git_artifact(id: nil)
      (@git_artifacts ||= {})[id]
    end

    def git_artifact_reset(id: nil)
      @git_artifacts.delete(id)
    end

    def git_artifact_filename(ending, id: nil)
      git_artifact(id: id).send(:filename, ending)
    end

    def git_artifact_tar_files(id: nil)
      shellout("tar -tf #{git_artifact_filename('.tar.gz', id: id)}").stdout.lines.map(&:strip).select { |f| !(f =~ %r{/$}) }
    end

    # rubocop:disable Metrics/AbcSize
    def git_artifact_archive(id: nil)
      git_artifact(id: id).archive_apply_command(stages[:source_1_archive])

      expect(@docker_image).to have_received(:add_git_artifact).with(
          %r{\/#{git_artifact_filename('.tar.gz', id: id)}$},
          git_artifact_filename('.tar.gz', id: id),
          git_artifact(id: id).where_to_add,
          step: :prepare
      )
      expect(File.read(git_artifact_filename('.commit', id: id)).strip).to eq(repo.latest_commit(git_artifact(id: id).branch))
      expect(File.exist?(git_artifact_filename('.tar.gz', id: id))).to be_truthy

      [:owner, :group].each do |subj|
        if git_artifact(id: id).send(subj)
          expect(send(:"tar_files_#{subj}s", git_artifact_filename('.tar.gz', id: id))).to eq([git_artifact(id: id).send(subj).to_s])
        end
      end
    end
    # rubocop:enable Metrics/AbcSize

    # rubocop:disable Metrics/AbcSize, Metrics/ParameterLists, Metrics/MethodLength
    def artifact_patch(suffix, step, id:, changefile: 'data.txt', changedata: random_string, should_be_empty: false, **_kwargs)
      repo_change_and_commit(changefile: changefile, changedata: changedata, branch: artifact(id: id).branch)

      reset_instances
      artifact(id: id).add_multilayer!

      patch_filename = artifact_filename("#{suffix}.patch.gz", id: id)
      patch_filename_esc = Regexp.escape(patch_filename)
      commit_filename = artifact_filename("#{suffix}.commit", id: id)

      if should_be_empty
        expect(@docker_image).to_not have_received(:add_artifact).with(/#{patch_filename_esc}$/, patch_filename, '/tmp', step: step)
        expect(@docker_image).to_not have_received(:run).with(/#{patch_filename_esc}/, /#{patch_filename_esc}$/, step: step)
        expect(File.exist?(patch_filename)).to be_falsy
        expect(File.exist?(commit_filename)).to be_falsy
      else
        expect(@docker_image).to have_received(:add_artifact).with(/#{patch_filename_esc}$/, patch_filename, '/tmp', step: step)
        expect(@docker_image).to have_received(:run).with(
            %r{^zcat \/tmp\/#{patch_filename_esc} \| .*git apply --whitespace=nowarn --directory=#{artifact(id: id).where_to_add}$},
            "rm /tmp/#{patch_filename}",
            step: step
        )
        { owner: 'u', group: 'g' }.each do |subj, flag|
          if artifact(id: id).send(subj)
            expect(@docker_image).to have_received(:run).with(/#{patch_filename_esc} \| sudo.*-#{flag} #{artifact(id: id).send(subj)}.*git apply/, any_args)
          end
        end
        expect(File.read(commit_filename).strip).to eq(repo.latest_commit(artifact(id: id).branch))
        expect(File.exist?(patch_filename)).to be_truthy
        expect(File.exist?(commit_filename)).to be_truthy
        expect(shellout("zcat #{patch_filename}").stdout).to match(/#{changedata}/)
      end
    end
    # rubocop:enable Metrics/AbcSize, Metrics/ParameterLists, Metrics/MethodLength
  end
end
