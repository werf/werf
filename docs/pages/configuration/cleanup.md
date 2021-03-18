---
title: Cleanup
sidebar: documentation
permalink: configuration/cleanup.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
summary: |
  <div class="language-yaml highlighter-rouge"><div class="highlight"><pre class="highlight"><code><span class="na">cleanup</span><span class="pi">:</span>
    <span class="na">keepPolicies</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="na">references</span><span class="pi">:</span>
        <span class="na">branch</span><span class="pi">:</span> <span class="s">&lt;string|/REGEXP/&gt;</span>
        <span class="na">tag</span><span class="pi">:</span> <span class="s">&lt;string|/REGEXP/&gt;</span>
        <span class="na">limit</span><span class="pi">:</span>
          <span class="na">last</span><span class="pi">:</span> <span class="s">&lt;int&gt;</span>
          <span class="na">in</span><span class="pi">:</span> <span class="s">&lt;duration string&gt;</span>
          <span class="na">operator</span><span class="pi">:</span> <span class="s">&lt;And|Or&gt;</span>
      <span class="na">imagesPerReference</span><span class="pi">:</span>
        <span class="na">last</span><span class="pi">:</span> <span class="s">&lt;int&gt;</span>
        <span class="na">in</span><span class="pi">:</span> <span class="s">&lt;duration string&gt;</span>
        <span class="na">operator</span><span class="pi">:</span> <span class="s">&lt;And|Or&gt;</span>
  </code></pre></div></div>  
---

## Configuring cleanup policies

The cleanup configuration consists of a set of policies called `keepPolicies`. They are used to select relevant images using the git history. Thus, during a [cleanup]({{ site.baseurl }}/reference/cleaning_process.html#git-history-based-cleanup-algorithm), __images not meeting the criteria of any policy are deleted__.

Each policy consists of two parts:

- `references` defines a set of references, git tags, or git branches to perform scanning on.
- `imagesPerReference` defines the limit on the number of images for each reference contained in the set.

Each policy should be linked to some set of git tags (`tag: <string|/REGEXP/>`) or git branches (`branch: <string|/REGEXP/>`). You can specify the name/group of a reference using the [Golang's regular expression syntax](https://golang.org/pkg/regexp/syntax/#hdr-Syntax).

```yaml
tag: v1.1.1
tag: /^v.*$/
branch: master
branch: /^(master|production)$/
```

> When scanning, werf searchs for the provided set of git branches in the origin remote references, but in the configuration, the  `origin/` prefix is omitted in branch names.

You can limit the set of references on the basis of the date when the git tag was created or the activity in the git branch. The `limit` group of parameters allows the user to define flexible and efficient policies for various workflows.

```yaml
- references:
    branch: /^features\/.*/
    limit:
      last: 10
      in: 168h
      operator: And
``` 

In the example above, werf selects no more than 10 latest branches that have the `features/` prefix in the name and have shown any activity during the last week.

- The `last: <int>` parameter allows you to select n last references from those defined in the `branch` / `tag`.
- The `in: <duration string>` parameter (you can learn more about the syntax in the [docs](https://golang.org/pkg/time/#ParseDuration)) allows you to select git tags that were created during the specified period or git branches that were active during the period. You can also do that for the specific set of `branches` / `tags`.
- The `operator: <And|Or>` parameter defines if references should satisfy both conditions or either of them (`And` is set by default).

When scanning references, the number of images is not limited by default. However, you can configure this behavior using the `imagesPerReference` set of parameters:

```yaml
imagesPerReference:
  last: <int>
  in: <duration string>
  operator: <And|Or>
```

- The `last: <int>` parameter defines the number of images to search for each reference. Their amount is unlimited by default (`-1`).
- The `in: <duration string>` parameter (you can learn more about the syntax in the [docs](https://golang.org/pkg/time/#ParseDuration)) defines the time frame in which werf searches for images.
- The `operator: <And|Or>` parameter defines what images will stay after applying the policy: those that satisfy both conditions or either of them (`And` is set by default).

> In the case of git tags, werf checks the HEAD commit only; the value of `last`>1 does not make any sense and is invalid

When describing a group of policies, you have to move from the general to the particular. In other words, `imagesPerReference` for a specific reference will match the latest policy it falls under:

```yaml
- references:
    branch: /.*/
  imagesPerReference:
    last: 1
- references:
    branch: master
  imagesPerReference:
    last: 5
```

In the above example, the _master_ reference matches both policies. Thus, when scanning the branch, the `last` parameter will equal to 5.

## Default policies

If there are no custom cleanup policies defined in `werf.yaml`, werf uses default policies configured as follows:

```yaml
cleanup:
  keepPolicies:
  - references:
      tag: /.*/
      limit:
        last: 10
  - references:
      branch: /.*/
      limit:
        last: 10
        in: 168h
        operator: And
    imagesPerReference:
      last: 2
      in: 168h
      operator: And
  - references:  
      branch: /^(master|staging|production)$/
    imagesPerReference:
      last: 10
``` 

Let us examine each policy individually:

1. Keep an image for the last 10 tags (by date of creation).
2. Keep no more than two images published over the past week, for no more than 10 branches active over the past week.
3. Keep the 10 latest images for master, staging, and production branches.
