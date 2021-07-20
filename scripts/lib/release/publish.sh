publish_binaries() {
    VERSION=$1

    echo "Publish version $VERSION from git tag $VERSION"

    docker run --rm -v $(pwd):/project --workdir /project -e AWS_ACCESS_KEY_ID=$S3_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY=$S3_SECRET_ACCESS_KEY amazon/aws-cli --endpoint-url $S3_ENDPOINT --region $S3_REGION s3 cp --recursive $RELEASE_BUILD_DIR/$VERSION s3://$S3_BUCKET_NAME/targets/releases/$VERSION
}
