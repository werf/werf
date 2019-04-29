create_release_message() {
    VERSION=$1

    TAG_TEMPLATE=scripts/lib/release/git_tag_template.md

    ( cat $TAG_TEMPLATE | VERSION=$VERSION envsubst | git tag --cleanup=verbatim --annotate --file - --edit $VERSION ) || ( return 1 )

    git push --tags
}
