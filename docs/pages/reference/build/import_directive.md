---
title: Importing artifacts
sidebar: reference
permalink: reference/build/import_directive.html
summary: |
  <a href="https://docs.google.com/drawings/d/e/2PACX-1vT9dsrIRkWKZaHNZG7g90JJHHHsAu3rxSh_5EWUWfkki3m0cQvIeUC2l01gRcYf0bGtxBLhvmcXn8d_/pub?w=2031&amp;h=144" data-featherlight="image">
    <img src="https://docs.google.com/drawings/d/e/2PACX-1vT9dsrIRkWKZaHNZG7g90JJHHHsAu3rxSh_5EWUWfkki3m0cQvIeUC2l01gRcYf0bGtxBLhvmcXn8d_/pub?w=1016&amp;h=72">
  </a>
    
  <div class="language-yaml highlighter-rouge"><pre class="highlight"><code><span class="s">import</span><span class="pi">:</span>
  <span class="pi">-</span> <span class="s">artifact</span><span class="pi">:</span> <span class="s">&lt;artifact name&gt;</span>
    <span class="s">before</span><span class="pi">:</span> <span class="s">&lt;install || setup&gt;</span>
    <span class="s">after</span><span class="pi">:</span> <span class="s">&lt;install || setup&gt;</span>
    <span class="s">add</span><span class="pi">:</span> <span class="s">&lt;absolute path&gt;</span>
    <span class="s">to</span><span class="pi">:</span> <span class="s">&lt;absolute path&gt;</span>
    <span class="s">owner</span><span class="pi">:</span> <span class="s">&lt;owner&gt;</span>
    <span class="s">group</span><span class="pi">:</span> <span class="s">&lt;group&gt;</span>
    <span class="s">includePaths</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;relative path or glob&gt;</span>
    <span class="s">excludePaths</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;relative path or glob&gt;</span></code></pre>
  </div>
---
