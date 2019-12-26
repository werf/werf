{% if include.header %}
{% assign header = include.header %}
{% else %}
{% assign header = "###" %}
{% endif %}

Retrieve a package from a package repository, and download it locally.

This is useful for fetching packages to inspect, modify, or repackage. It can
also be used to perform cryptographic verification of a chart without installing
the chart.

There are options for unpacking the chart after download. This will create a
directory for the chart and uncompress into that directory.

If the --verify flag is specified, the requested chart MUST have a provenance
file, and MUST pass the verification process. Failure in any part of this will
result in an error, and the chart will not be saved locally.


{{ header }} Syntax

```shell
werf helm repo fetch [chart URL | repo/chartname] [...] [options]
```

{{ header }} Options

```shell
      --ca-file='':
            verify certificates of HTTPS-enabled servers using this CA bundle
      --cert-file='':
            identify HTTPS client using this SSL certificate file
  -d, --destination='.':
            location to write the chart. If this and tardir are specified, tardir is appended to    
            this
      --devel=false:
            use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set,   
            this is ignored.
      --helm-home='~/.helm':
            location of your Helm config. Defaults to $WERF_HELM_HOME, $HELM_HOME or ~/.helm
  -h, --help=false:
            help for fetch
      --key-file='':
            identify HTTPS client using this SSL key file
      --keyring='$HOME/.gnupg/pubring.gpg':
            keyring containing public keys
      --password='':
            chart repository password
      --prov=false:
            fetch the provenance file, but don't perform verification
      --repo='':
            chart repository url where to locate the requested chart
      --untar=false:
            if set to true, will untar the chart after downloading it
      --untardir='.':
            if untar is specified, this flag specifies the name of the directory into which the     
            chart is expanded
      --username='':
            chart repository username
      --verify=false:
            verify the package against its signature
      --version='':
            specific version of a chart. Without this, the latest version is fetched
```

