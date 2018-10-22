module Dapp
  module Dimg
    # Git repo artifact
    class GitArtifact
      include Helper::Tar
      include Helper::Trivia

      attr_reader :repo
      attr_reader :name
      attr_reader :as

      # FIXME: переименовать cwd в from

      # rubocop:disable Metrics/ParameterLists
      def initialize(repo, dimg, to:, name: nil, branch: nil, tag: nil, commit: nil,
                     cwd: nil, include_paths: nil, exclude_paths: nil, owner: nil, group: nil, as: nil,
                     stages_dependencies: {}, ignore_signature_auto_calculation: false)
        @repo = repo
        @dimg = dimg
        @name = name

        @ignore_signature_auto_calculation = ignore_signature_auto_calculation

        @branch = branch
        @tag = tag
        @commit = commit

        @to = to
        @cwd = (cwd.nil? || cwd.empty? || cwd == '/') ? '' : File.expand_path(File.join('/', cwd, '/'))[1..-1]
        @include_paths = include_paths
        @exclude_paths = exclude_paths
        @owner = owner
        @group = group
        @as = as
        @stages_dependencies = stages_dependencies
      end
      # rubocop:enable Metrics/ParameterLists

      def apply_archive_command(stage)
        res = repo.dapp.ruby2go_git_artifact(
          "GitArtifact" => JSON.dump(get_ruby2go_state_hash),
          "method" => "ApplyArchiveCommand",
          "Stage" => JSON.dump(merge_layer_commit_stub_stage_state(stage.get_ruby2go_state_hash, stage)),
        )

        raise res["error"] if res["error"]

        self.set_ruby2go_state_hash(JSON.load(res["data"]["GitArtifact"]))
        stage.set_ruby2go_state_hash(JSON.load(res["data"]["Stage"]))

        Array(res["data"]["result"])
      end

      def merge_layer_commit_stub_stage_state(stage_state, stage)
        # Data for StubStage specific for ApplyPatchCommand
        stage_state["LayerCommitMap"] = {
          paramshash => stage.layer_commit(self),
        }
        stage_state["PrevStage"]["LayerCommitMap" ] = {
          paramshash => stage.prev_stage.layer_commit(self),
        }

        stage_state
      end

      def apply_patch_command(stage)
        res = repo.dapp.ruby2go_git_artifact(
          "GitArtifact" => JSON.dump(get_ruby2go_state_hash),
          "method" => "ApplyPatchCommand",
          "Stage" => JSON.dump(merge_layer_commit_stub_stage_state(stage.get_ruby2go_state_hash, stage)),
        )

        raise res["error"] if res["error"]

        self.set_ruby2go_state_hash(JSON.load(res["data"]["GitArtifact"]))
        stage.set_ruby2go_state_hash(JSON.load(res["data"]["Stage"]))

        Array(res["data"]["result"])
      end

      def calculate_stage_dependencies_checksum(stage)
        res = repo.dapp.ruby2go_git_artifact(
          "GitArtifact" => JSON.dump(get_ruby2go_state_hash),
          "method" => "StageDependenciesChecksum",
          "StageName" => ::Dapp::Helper::CaseConversion.snake_case_to_lower_camel_case(stage.name.to_s),
        )

        raise res["error"] if res["error"]

        self.set_ruby2go_state_hash(JSON.load(res["data"]["GitArtifact"]))

        result = res["data"]["result"]
        return [] if result == ""
        return result
      end

      def stage_dependencies_checksum(stage)
        stage_dependencies_key = [stage.name, commit]
        @stage_dependencies_checksums ||= {}
        @stage_dependencies_checksums[stage_dependencies_key] ||= calculate_stage_dependencies_checksum(stage)
      end

      def patch_size(from_commit)
        res = repo.dapp.ruby2go_git_artifact(
          "GitArtifact" => JSON.dump(get_ruby2go_state_hash),
          "method" => "PatchSize",
          "FromCommit" => from_commit,
        )

        raise res["error"] if res["error"]

        self.set_ruby2go_state_hash(JSON.load(res["data"]["GitArtifact"]))

        res["data"]["result"]
      end

      def get_ruby2go_state_hash
        {
          "Name" => @name.to_s,
          "As" => @as.to_s,
          "Branch" => @branch.to_s,
          "Tag" => @tag.to_s,
          "Commit" => @commit.to_s,
          "To" => @to.to_s,
          "Cwd" => @cwd.to_s,
          "RepoPath" => File.join("/", @cwd.to_s),
          "Owner" => @owner.to_s,
          "Group" => @group.to_s,
          "IncludePaths" => @include_paths,
          "ExcludePaths" => @exclude_paths,
          "StagesDependencies" => @stages_dependencies.map {|k, v| [_stages_map[k], Array(v).map(&:to_s)]}.to_h,
          "PatchesDir" => dimg.tmp_path('patches'),
          "ContainerPatchesDir" => dimg.container_tmp_path('patches'),
          "ArchivesDir" => dimg.tmp_path('archives'),
          "ContainerArchivesDir" => dimg.container_tmp_path('archives'),
        }.tap {|res|
          if repo.is_a? ::Dapp::Dimg::GitRepo::Local
            res["LocalGitRepo"] = repo.get_ruby2go_state_hash
          elsif repo.is_a? ::Dapp::Dimg::GitRepo::Remote
            res["RemoteGitRepo"] = repo.get_ruby2go_state_hash
          else
            raise
          end
        }
      end

      def _stages_map
        {
          before_install: "beforeInstall",
          install: "install",
          before_setup: "beforeSetup",
          setup: "setup",
          build_artifact: "buildArtifact",
        }
      end

      def _stages_map_reversed
        _stages_map.map {|k, v| [v, k]}.to_h
      end

      def set_ruby2go_state_hash(new_state)
        [
          [:@name, new_state["Name"]],
          [:@as, new_state["As"]],
          [:@branch, new_state["Branch"]],
          [:@tag, new_state["Tag"]],
          [:@commit, new_state["Commit"]],
          [:@cwd, new_state["Cwd"]],
          [:@owner, new_state["Owner"]],
          [:@group, new_state["Group"]],
        ].each do |var, new_value|
          if new_value != ""
            instance_variable_set(var, new_value)
          end
        end

        @stages_dependencies = new_state["StagesDependencies"].map do |k, v|
          [_stages_map_reversed[k], v]
        end.to_h
      end

      def latest_commit
        @latest_commit ||= begin
          res = repo.dapp.ruby2go_git_artifact("GitArtifact" => JSON.dump(get_ruby2go_state_hash), "method" => "LatestCommit")

          raise res["error"] if res["error"]

          self.set_ruby2go_state_hash(JSON.load(res["data"]["GitArtifact"]))

          res["data"]["result"]
        end.tap do |c|
          repo.dapp.log_info("Repository `#{repo.name}`: latest commit `#{c}` to `#{to}`") unless ignore_signature_auto_calculation
        end
      end

      def paramshash
        @paramshash ||= begin
          res = repo.dapp.ruby2go_git_artifact("GitArtifact" => JSON.dump(get_ruby2go_state_hash), "method" => "GetParamshash")

          raise res["error"] if res["error"]

          self.set_ruby2go_state_hash(JSON.load(res["data"]["GitArtifact"]))

          res["data"]["result"]
        end
      end

      def full_name
        @full_name ||= begin
          res = repo.dapp.ruby2go_git_artifact("GitArtifact" => JSON.dump(get_ruby2go_state_hash), "method" => "FullName")

          raise res["error"] if res["error"]

          self.set_ruby2go_state_hash(JSON.load(res["data"]["GitArtifact"]))

          res["data"]["result"]
        end
      end

      def is_patch_empty(stage)
        return @is_patch_empty if !@is_patch_empty.nil?

        @is_patch_empty = begin
          res = repo.dapp.ruby2go_git_artifact(
            "GitArtifact" => JSON.dump(get_ruby2go_state_hash),
            "method" => "IsPatchEmpty",
            "Stage" => JSON.dump(merge_layer_commit_stub_stage_state({"PrevStage" => {}}, stage)),
          )

          raise res["error"] if res["error"]

          self.set_ruby2go_state_hash(JSON.load(res["data"]["GitArtifact"]))

          res["data"]["result"]
        end
      end

      def is_empty
        return @is_empty if !@is_empty.nil?

        @is_empty = begin
          res = repo.dapp.ruby2go_git_artifact(
            "GitArtifact" => JSON.dump(get_ruby2go_state_hash),
            "method" => "IsEmpty",
          )

          raise res["error"] if res["error"]

          self.set_ruby2go_state_hash(JSON.load(res["data"]["GitArtifact"]))

          res["data"]["result"]
        end
      end

      protected

      attr_reader :dimg
      attr_reader :to
      attr_reader :commit
      attr_reader :branch
      attr_reader :tag
      attr_reader :cwd
      attr_reader :owner
      attr_reader :group
      attr_reader :stages_dependencies
      attr_reader :ignore_signature_auto_calculation
    end
  end
end
