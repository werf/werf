---
title: Stage Introspection
sidebar: reference
permalink: reference/build/stage_introspection.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
summary: |
  <div class="language-bash highlighter-rouge"><pre class="highlight"><code><span class="c"># introspection before and after execution of a dysfunctional set of instructions</span>
  werf dimg build --introspect-error 
  werf dimg build --introspect-before-error
  
  <span class="c"># introspection of an assembled STAGE stage</span>
  werf dimg build --introspect-stage STAGE
  werf dimg build --introspect-artifact-stage STAGE
  
  <span class="c"># introspection before execution of the instructions for the STAGE stage</span>
  werf dimg build --introspect-before STAGE
  werf dimg build --introspect-artifact-before STAGE   
  </code></pre>
  </div>
---

Writing a configuration is especially difficult at the beginning because you don't quite understand what is in the _stage assembly container_ when the instructions are executed.

In the process of assembling, you can access a certain _stage_ using introspection options. During introspection, like during assembling, the _stage assembly container_ contains service tools and environment variables. Tools are presented as a set of utilities required during assembling. They are added by mounting the directories from service containers of our _werfdeps_ distributions (available at `/.werf/deps` path in the _assembly container_). Introspection comes down the fact that the _stage assembly container_ is launched for users in interactive mode.

**During development**, introspection makes it possible to achieve the required outcomes in an _assembly container_, and then transfer all the steps and instructions into the configuration of the appropriate _stage_. This approach is useful when the set objective is clear, although the steps to achieve it are not so obvious and require a great deal of experiment.

<div class="videoWrapper">
<iframe width="560" height="315" src="https://www.youtube.com/embed/quoWwLSM_-4" frameborder="0" allow="encrypted-media" allowfullscreen></iframe>
</div>
                  
**During debugging**, introspection allows you to see why assembling ended with an error or the result was unexpected, to check whether dependent files are present and to check the system state.

<div class="videoWrapper">
<iframe width="560" height="315" src="https://www.youtube.com/embed/GiEbEhF2Pes" frameborder="0" allow="encrypted-media" allowfullscreen></iframe>
</div>

Finally, when introspection is effected for applications that use **ansible**,  you can debug ansible playbooks in the _assembly container_ and subsequently transfer ansible tasks to the appropriate configuration _stages_.

<div class="videoWrapper">
<iframe width="560" height="315" src="https://www.youtube.com/embed/TEpn0yFvJik" frameborder="0" allow="encrypted-media" allowfullscreen></iframe>
</div>
