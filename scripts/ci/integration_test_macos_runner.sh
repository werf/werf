#!/bin/bash -ex

if [[ -z "$GITHUB_REF" ]] || [[ -z "$GITHUB_SHA" ]]; then
  echo "script requires \$GITHUB_REF and \$GITHUB_SHA" >&2
  exit 1
fi

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

tests_tag=${1:integration}

job_name=${JOB_NAME:-integration_tests}
workflow_ref=$(echo -n $GITHUB_REF | tr / _)
workflow_job_filename=${job_name}_${workflow_ref}

job_lockfile=~/.github_actions/lock/job/${workflow_ref}
job_pgidfile=~/.github_actions/pgid/${workflow_job_filename}
workflow_job_lockfile=~/.github_actions/lock/workflow/${workflow_job_filename}
workflow_job_commit_file=~/.github_actions/commit/${workflow_job_filename}

mkdir -p $(dirname $job_lockfile) $(dirname $job_pgidfile) $(dirname $workflow_job_lockfile) $(dirname $workflow_job_commit_file)

(
    flock 9

    if [[ -f $job_pgidfile ]]; then
      pgid_id=$(cat $job_pgidfile)
      ps -p $pgid_id || ecode=$?

      if [[ "$ecode" -eq 0 ]];then
        current_commit_ts=$(git show -s --format=%ct $GITHUB_SHA)
        if [[ -f $workflow_job_commit_file ]]; then
          commit_sha=$(cat $workflow_job_commit_file)
          saved_commit_ts=$(git show -s --format=%ct ${commit_sha} || echo 0)
        fi

        if [[ "$current_commit_ts" -gt "$saved_commit_ts" ]]; then
          kill -- -$pgid_id || true
        else
          echo "Test was skipped"
          exit 1
        fi
      fi
    fi

    pgid=$(ps -o pgid= $$ | grep -o [0-9]*)
    echo $pgid > $job_pgidfile
    echo $GITHUB_SHA > $workflow_job_commit_file
) 9>$workflow_job_lockfile

get_lock_start_ts=$(date +%s)
(
    flock 9

    get_lock_end_ts=$(date +%s)
    echo "Got lock in $((get_lock_end_ts-$get_lock_start_ts)) seconds"

    if [[ "$tests_tag" == "integration_k8s" ]]; then
        source ${script_dir}/integration_k8s_tests_before_hook.sh
        exec ginkgo --tags integration_k8s -p -r integration
    else
        exec ginkgo --tags integration -p -r integration
    fi
) 9>$job_lockfile
