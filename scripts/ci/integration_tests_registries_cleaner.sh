#!/bin/bash -ex

if [[ -z "$1" ]]; then
  echo "script requires argument <implementation name>" >&2
  exit 1
fi

name=${1^^}

case $name in

ACR)
  az acr repository list --name werftests | jq .[] | xargs -r -n1 sh -c 'echo $0; az acr repository delete --name=werftests --repository=$0 --yes || exit 255'
  ;;

ECR)
  aws ecr describe-repositories | jq '.repositories[].repositoryName' | tr -d '"' | xargs -r -n1 sh -c 'echo $0; aws ecr delete-repository --repository-name $0 --force || exit 255'
  ;;

DOCKERHUB)
  username=$WERF_TEST_DOCKERHUB_USERNAME
  password=$WERF_TEST_DOCKERHUB_PASSWORD

  if [[ -z "$username" ]] || [[ -z "$password" ]]; then
    echo "script requires specified \$WERF_TEST_DOCKERHUB_USERNAME and \$WERF_TEST_DOCKERHUB_PASSWORD" >&2
    exit 1
  fi

  token=$(curl -s -H "Content-Type: application/json" -X POST -d '{"username": "'${username}'", "password": "'${password}'"}' https://hub.docker.com/v2/users/login/ | jq -r .token)

  while :
  do
    repo_list=($(curl -s -H "Authorization: JWT ${token}" https://hub.docker.com/v2/repositories/${username}/?page_size=100 | jq -r '.results|.[]|.name'))

    if [ ${#repo_list[@]} -eq 0 ]; then
      break
    fi

    for repo in ${repo_list[@]}
    do
      curl -X DELETE -s -H "Authorization: JWT ${token}" https://hub.docker.com/v2/repositories/${username}/${repo}/
    done
  done
  ;;

GCR)
  delete_untagged() {
    while read digest; do
       gcloud container images delete $1@$digest --quiet 2>&1 | sed 's/^/        /'
    done < <(gcloud container images list-tags $1 --filter='-tags:*' --format='get(digest)' --limit=unlimited)
  }

  delete_for_each_repo() {
    while read repo; do
        delete_untagged $repo
    done < <(gcloud container images list --repository $1 --format="value(name)")
  }

  project_id=$GOOGLE_PROJECT_ID
  if [[ -z "$project_id" ]]; then
    echo "script requires specified \$GOOGLE_PROJECT_ID" >&2
    exit 1
  fi

  delete_for_each_repo gcr.io/"$project_id"
  ;;

QUAY)
  username=$WERF_TEST_QUAY_USERNAME
  token=$WERF_TEST_QUAY_TOKEN

  if [[ -z "$username" ]] || [[ -z "$token" ]]; then
    echo "script requires specified \$WERF_TEST_QUAY_USERNAME and \$WERF_TEST_QUAY_TOKEN" >&2
    exit 1
  fi

  docker run \
    --rm \
    -t \
    --entrypoint=sh \
    -e QUAY_API_TOKEN=${token} \
    -e QUAY_HOSTNAME=quay.io \
    quay.io/koudaiii/qucli:latest -c "/qucli list $username --is-public=false | grep werf | awk '{print \$1}' | sed 's/quay.io\///' | xargs -r -n1 sh -c 'echo \$0; /qucli delete \$0 || exit 255'"
  ;;

*)
  echo "implementation name $name is not supported" >&2
  exit 1
  ;;
esac
