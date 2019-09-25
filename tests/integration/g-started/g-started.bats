setup() {
	echo "Setup begin!" >&3
	export SAVE_SETUP_PWD=$(pwd)
	echo "Setup done!" >&3
}

teardown() {
	echo "Teardown begin!" >&3

	werf stages purge --stages-storage :local >&3

	cd $SAVE_SETUP_PWD

	rm -rf ./g-started
	docker rm -f registry g-started

	echo "Teardown done!" >&3
}

@test "g-started" {
	git clone https://github.com/dockersamples/linux_tweet_app.git g-started

	cd g-started
	cat <<EOF > werf.yaml
project: g-started
configVersion: 1
---
image: ~
dockerfile: Dockerfile
EOF

	werf build --stages-storage :local >&3
	werf run --stages-storage :local --docker-options="-d -p 80:80 --name g-started"
	curl localhost:80
	docker run -d -p 5000:5000 --restart=always --name registry registry:2
	werf publish --stages-storage :local --images-repo localhost:5000/g-started --tag-custom v0.1.0 >&3
}
