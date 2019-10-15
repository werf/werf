load ../../../../helpers/common

setup() {
    werf_home_init
    test_dir_create
    test_dir_cd
}

teardown() {
    docker rm --force \
        $PAYMENT_GW_CONTAINER_NAME \
        $DATABASE_CONTAINER_NAME \
        $APP_CONTAINER_NAME \
        $REVERSE_PROXY_CONTAINER_NAME

    test_dir_werf_stages_purge
    test_dir_rm
    werf_home_deinit
}

@test "multi images" {
    git clone https://github.com/dockersamples/atsea-sample-shop-app.git .
    mkdir -p reverse_proxy/certs && openssl req -newkey rsa:4096 -nodes -subj "/CN=atseashop.com;" -sha256 -keyout reverse_proxy/certs/revprox_key -x509 -days 365 -out reverse_proxy/certs/revprox_cert
    cp $BATS_TEST_DIRNAME/data/werf.yaml .

    werf build \
        --stages-storage :local

    PAYMENT_GW_CONTAINER_NAME=payment_gw-$(generate_random_string)
    werf run \
        --stages-storage :local \
        --docker-options="--rm -d --name $PAYMENT_GW_CONTAINER_NAME" \
        payment_gw

    DATABASE_CONTAINER_NAME=database-$(generate_random_string)
    database_host_port=$(get_unused_port)
    werf run \
        --stages-storage :local \
        --docker-options="--rm -d -p $database_host_port:5432 --name $DATABASE_CONTAINER_NAME" \
        database

    APP_CONTAINER_NAME=app-$(generate_random_string)
    app_host_port=$(get_unused_port)
    werf run \
        --stages-storage :local \
        --docker-options="--rm -d -p $app_host_port:8080 --link $DATABASE_CONTAINER_NAME:database --name $APP_CONTAINER_NAME" \
        app

    REVERSE_PROXY_CONTAINER_NAME=reverse-proxy-$(generate_random_string)
    reverse_proxy_container_host_port80=$(get_unused_port)
    reverse_proxy_container_host_port443=$(get_unused_port)
    werf run \
        --stages-storage :local \
        --docker-options="--rm -d -p $reverse_proxy_container_host_port80:80 -p $reverse_proxy_container_host_port443:443 --link $APP_CONTAINER_NAME:appserver --name $REVERSE_PROXY_CONTAINER_NAME" \
        reverse_proxy

    wait_till_host_ready_to_respond localhost:$app_host_port 60
    run curl localhost:$app_host_port/index.html
    [ "$status" -eq 0 ]
    [[ "$output" =~ "Atsea Shop" ]]

    docker stop \
        $PAYMENT_GW_CONTAINER_NAME \
        $DATABASE_CONTAINER_NAME \
        $APP_CONTAINER_NAME \
        $REVERSE_PROXY_CONTAINER_NAME
}
