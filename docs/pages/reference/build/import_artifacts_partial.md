<div class="summary" markdown="1">

<a href="https://docs.google.com/drawings/d/e/2PACX-1vT9dsrIRkWKZaHNZG7g90JJHHHsAu3rxSh_5EWUWfkki3m0cQvIeUC2l01gRcYf0bGtxBLhvmcXn8d_/pub?w=2031&amp;h=144" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vT9dsrIRkWKZaHNZG7g90JJHHHsAu3rxSh_5EWUWfkki3m0cQvIeUC2l01gRcYf0bGtxBLhvmcXn8d_/pub?w=1016&amp;h=72">
</a>

```yaml
import:
- artifact: <artifact name>
  before: <install || setup>
  after: <install || setup>
  add: <absolute path>
  to: <absolute path>
  owner: <owner>
  group: <group>
  includePaths:
  - <relative path or glob>
  excludePaths:
  - <relative path or glob>
```

</div>

Importing _artifact resources_ into an image is described using the `import` directive, which is an array of import records. Each record should contain the following fields:
 
- `artifact: <artifact name>` — the name of the _artifact_ from which you want to copy the files
- `add: <absolute path>` — the absolute path in the _artifact image_ to the file or folder for copying
- `to: <absolute path>` — the absolute path in the destination _image_, where the resources from the _artifact image_ should be copied. In case of absence, it is equal to the value of the directive `add`
- `before: <install || setup>` or `after: <install || setup>` — to specify stage where import the artifact files into the image. At present
only _install_ and _setup_ stages are supported

```yaml
import:
- artifact: application-assets
  add: /app/public/assets
  after: install
- artifact: application-assets
  add: /app/vendor
  after: install
```

As in the case of adding _git paths_, masks are supported for including and excluding files from the specified path, and you can also specify the rights for the imported resources. Read more about these in the [git directive article]({{ site.baseurl }}/reference/build/git_directive.html).

> Import paths and _git paths_ should not overlap with each other.
