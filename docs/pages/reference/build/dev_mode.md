---
title: Work with local changes without commits in dev mode
sidebar: reference
permalink: reference/build/dev_mode.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
summary: |
  <div class="language-bash highlighter-rouge"><pre class="highlight"><code>dapp dimg build --dev
  
  <span class="c"># removing dev cache for the stages of all projects</span>
  dapp dimg mrproper --improper-dev-mode-cache    
  </code></pre>
  </div>
---

When developing the configuration in standard assembly mode, formal commits are required so that dapp can take account of changing in files while assembling. Similar to mounting a working directory during Dockerfile assembly, you may want to work with the current state of the local repository.

To simplify and accelerate the assembly configuration development workflow, dapp provides a special _developer mode_ that is enabled by `--dev` option. Special features of _developer mode_:
* During assembly, non-committed changes of the local git repository that comply with the configuration are considered. More specifically, the paths from the following _git-path_ sub-directives are taken into account: `add`, `includePaths`, `excludePaths`, `stageDependencies`.
* Successfully assembled _stages_ always saves to the _stages cache_ (more information about _stages cache_ can be found in the [relevant article]({{ site.baseurl }}/reference/build/cache.html#forced-saving-images-to-cache-after-assembling)).
* Dapp creates a separate _stages cache_ that in no way impacts the standard one.

Working in _developer mode_ may be useful not only for local development but also when you need to debug your code on the build machine without impacting the assembly cache or making commits to the git repository.
