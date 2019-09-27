test_case_run() {
  werf build --stages-storage :local

  container_name=$1
  container_host_port=$(get_unused_port)
  werf run \
    --stages-storage :local \
    --docker-options="--rm -d -p $container_host_port:8000 --name $container_name" -- /app/start.sh

  wait_till_host_ready_to_respond localhost:$container_host_port
  run curl localhost:$container_host_port
  [ "$status" -eq 0 ]
  [[ "$output" =~ "Symfony Demo application" ]]

  registry_repository_name=$container_name
  werf publish \
    --stages-storage :local \
    --images-repo $WERF_TEST_DOCKER_REGISTRY/$registry_repository_name \
    --tag-custom v0.1.0

  docker stop $container_name
}
