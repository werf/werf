module SpecHelper
  module Git
    extend ActiveSupport::Concern

    def git_init!(git_dir: nil)
      FileUtils.mkdir_p git_dir unless git_dir.nil?
      git('init', git_dir: git_dir)
      git('commit --allow-empty -m init', git_dir: git_dir)
    end

    def git_create_branch!(branch, git_dir: nil)
      git("branch #{branch}", git_dir: git_dir)
    end

    def git_change_and_commit!(changefile = 'data.txt', changedata = random_string, branch: 'master', git_dir: nil)
      git("checkout #{branch}", git_dir: git_dir)
      changefile = File.join([git_dir, changefile].compact)
      FileUtils.mkdir_p File.split(changefile)[0]
      File.write changefile, changedata
      git_commit!(git_dir: git_dir)
    end

    def git_commit!(git_dir: nil)
      git('add --all', git_dir: git_dir)
      git('commit -m +', git_dir: git_dir) unless git('diff --cached --quiet', returns: [0, 1], git_dir: git_dir).status.success?
    end

    def git_latest_commit(git_dir: nil, branch: 'master')
      git("rev-parse #{branch}", git_dir: git_dir).stdout.strip
    end

    def git(command, git_dir: nil, **kwargs)
      commands = [
        "#{dimg.project.git_bin} config -l | grep 'user.email' || #{dimg.project.git_bin} config --global user.email 'dapp@flant.com'",
        "#{dimg.project.git_bin} config -l | grep 'user.name' || #{dimg.project.git_bin} config --global user.name 'Dapp Dapp'",
        "#{dimg.project.git_bin} #{"-C #{git_dir}" unless git_dir.nil?} #{command}"
      ]

      dimg.project.system_shellout_extra(volume: @test_dir) do
        dimg.system_shellout!(commands.join(' && '), **kwargs)
      end
    end
  end
end
