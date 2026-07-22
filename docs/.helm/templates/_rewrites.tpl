# SHOULD BE IN SYNC WITH github.com/werf/website/.helm/templates/_rewrites.tpl
{{- define "rewrites" }}
{{- $currentRoot := .Values.docs.currentRoot }}
{{- $supportedRoots := .Values.docs.supportedRoots }}
{{- $supportedRootPatterns := list }}
{{- range $root := $supportedRoots }}
{{- $supportedRootPatterns = append $supportedRootPatterns (replace "." "\\." $root) }}
{{- end }}
{{- $supportedRootsPattern := join "|" $supportedRootPatterns }}
{{- $supportedOrLatestPattern := $supportedRootsPattern }}
{{- if .Values.docs.latestAliasEnabled }}
{{- $supportedOrLatestPattern = printf "latest|%s" $supportedOrLatestPattern }}
{{- end }}
{{- $addressableDocsPattern := printf "pr-[^/]+|%s" $supportedOrLatestPattern }}

############################################
# Normalize urls
############################################

rewrite ^/js/(?<tail>.+)                                                               /assets/js/$tail        redirect;
rewrite ^/css/(?<tail>.+)                                                              /assets/css/$tail       redirect;
rewrite ^/images/(?<tail>.+)                                                           /assets/images/$tail    redirect;

rewrite ^/docs\.html$                                                                  /docs/                  redirect;
rewrite ^/docs/(?<ver>{{ $addressableDocsPattern }})$                                  /docs/$ver/             redirect;
rewrite ^/docs/(?<ver>{{ $addressableDocsPattern }})/index\.html$                      /docs/$ver/             redirect;
rewrite ^/(?<ver>{{ $addressableDocsPattern }})/docs\.html$                            /docs/$ver/             redirect;
rewrite ^/(?<ver>{{ $addressableDocsPattern }})/docs/?$                                 /docs/$ver/             redirect;
rewrite ^/(?<ver>{{ $addressableDocsPattern }})/docs/(?<tail>.+)                        /docs/$ver/$tail        redirect;

rewrite ^/documentation\.html$                                                         /docs/                  redirect;
rewrite ^/documentation/?$                                                             /docs/                  redirect;
rewrite ^/documentation/(?<ver>{{ $addressableDocsPattern }})/?$                       /docs/$ver/             redirect;
rewrite ^/documentation/(?<ver>{{ $addressableDocsPattern }})/index\.html$            /docs/$ver/             redirect;
rewrite ^/documentation/(?<ver>{{ $addressableDocsPattern }})/(?<tail>.+)              /docs/$ver/$tail        redirect;
rewrite ^/(?<ver>{{ $addressableDocsPattern }})/documentation\.html$                   /docs/$ver/             redirect;
rewrite ^/(?<ver>{{ $addressableDocsPattern }})/documentation/?$                        /docs/$ver/             redirect;
rewrite ^/(?<ver>{{ $addressableDocsPattern }})/documentation/(?<tail>.+)               /docs/$ver/$tail        redirect;

############################################
# Docs roots and shortcuts
############################################

rewrite ^/docs/?$                                                                           /docs/{{ $currentRoot }}/                                       redirect;
{{- range $root := $supportedRoots }}
{{- $rootPattern := replace "." "\\." $root }}
rewrite ^/docs/(?<ver>{{ $rootPattern }}\.[^/]+)/(?<tail>.*)                              /docs/{{ $root }}/$tail                                         redirect;
rewrite ^/(?<ver>{{ $rootPattern }}\.[^/]+)/docs\.html$                                    /docs/{{ $root }}/                                              redirect;
rewrite ^/(?<ver>{{ $rootPattern }}\.[^/]+)/docs/?$                                         /docs/{{ $root }}/                                              redirect;
rewrite ^/(?<ver>{{ $rootPattern }}\.[^/]+)/docs/(?<tail>.+)                                /docs/{{ $root }}/$tail                                         redirect;
rewrite ^/documentation/(?<ver>{{ $rootPattern }}\.[^/]+)/?$                                /docs/{{ $root }}/                                              redirect;
rewrite ^/documentation/(?<ver>{{ $rootPattern }}\.[^/]+)/index\.html$                     /docs/{{ $root }}/                                              redirect;
rewrite ^/documentation/(?<ver>{{ $rootPattern }}\.[^/]+)/(?<tail>.+)                       /docs/{{ $root }}/$tail                                         redirect;
rewrite ^/(?<ver>{{ $rootPattern }}\.[^/]+)/documentation\.html$                            /docs/{{ $root }}/                                              redirect;
rewrite ^/(?<ver>{{ $rootPattern }}\.[^/]+)/documentation/?$                                 /docs/{{ $root }}/                                              redirect;
rewrite ^/(?<ver>{{ $rootPattern }}\.[^/]+)/documentation/(?<tail>.+)                        /docs/{{ $root }}/$tail                                         redirect;
{{- end }}
rewrite ^/docs/(?!(?:{{ $addressableDocsPattern }})/)(?:.+)                                /docs/{{ $currentRoot }}/                                       redirect;

rewrite ^/docs/(?<ver>{{ $supportedOrLatestPattern }})/?$                                      /docs/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/docs/(?<ver>pr-[^/]+)/?$                                                              /docs/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/docs/(?<ver>{{ $supportedOrLatestPattern }})/usage/?$                                /docs/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/docs/(?<ver>{{ $supportedOrLatestPattern }})/usage/project_configuration/?$          /docs/$ver/usage/project_configuration/overview.html            redirect;
rewrite ^/docs/(?<ver>{{ $supportedOrLatestPattern }})/usage/build/?$                          /docs/$ver/usage/build/overview.html                            redirect;
rewrite ^/docs/(?<ver>{{ $supportedOrLatestPattern }})/usage/build/stapel/?$                   /docs/$ver/usage/build/stapel/overview.html                     redirect;
rewrite ^/docs/(?<ver>{{ $supportedOrLatestPattern }})/usage/deploy/?$                         /docs/$ver/usage/deploy/overview.html                           redirect;
rewrite ^/docs/(?<ver>{{ $supportedOrLatestPattern }})/usage/distribute/?$                     /docs/$ver/usage/distribute/overview.html                       redirect;
rewrite ^/docs/(?<ver>{{ $supportedOrLatestPattern }})/usage/cleanup/?$                        /docs/$ver/usage/cleanup/cr_cleanup.html                        redirect;
rewrite ^/docs/(?<ver>{{ $supportedOrLatestPattern }})/reference/?$                            /docs/$ver/reference/werf_yaml.html                             redirect;
rewrite ^/docs/(?<ver>{{ $supportedOrLatestPattern }})/reference/cli/?$                        /docs/$ver/reference/cli/overview.html                          redirect;
rewrite ^/docs/(?<ver>{{ $supportedOrLatestPattern }})/resources/?$                            /docs/$ver/resources/cheat_sheet.html                           redirect;


############################################
# Redirects for moved or deleted urls
############################################

rewrite ^/installation\.html$                                                                     /getting_started/                                      redirect;
rewrite ^/applications_guide_(?:ru|en)/?                                                          /guides.html                                           redirect;
rewrite ^/publications_ru\.html$                                                                  https://ru.werf.io/publications.html                   redirect;
rewrite ^/how_it_works\.html                                                                      /#how-it-works                                         redirect;
rewrite ^/introduction\.html$                                                                     /#how-it-works                                         redirect;

{{- end }}
