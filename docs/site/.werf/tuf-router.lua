local dst_url
local res

ngx.var.arch = string.lower(ngx.var.arch)
ngx.var.arch = string.gsub(ngx.var.arch, 'x86_64', 'amd64')
ngx.var.arch = string.gsub(ngx.var.arch, 'aarch64', 'arm64')
local m, err = ngx.re.match(ngx.var.target, '.sig$')

if m then
   dst_url = '/targets/signature/releases/'
else
   dst_url = '/targets/releases/'
end

local version_request_url = string.format('/targets/channels/%s/%s', ngx.var.group, ngx.var.channel)

-- A chain of redirects is not working well, so, avoid it here
local max_hop = 1
repeat
  res = ngx.location.capture(version_request_url)
  if res.status == ngx.HTTP_MOVED_PERMANENTLY or res.status == ngx.HTTP_MOVED_TEMPORARILY then
    ngx.log(ngx.NOTICE, string.format('Got redirect %s to %s (requested - %s)', res.status, res.header['Location'], version_request_url))
    version_request_url = res.header['Location']
  end
  max_hop = max_hop - 1
until not ( max_hop > 0 and res.status ~= ngx.HTTP_OK )

if res.status == ngx.HTTP_OK then
  -- Is this enough to validate the version?
  ngx.var.version = string.gsub(res.body,'[^a-zA-Z0-9.+-]','')
  ngx.header["Cache-Control"] = 'no-store, no-cache'
  return ngx.redirect(string.format('%s%s%s/%s-%s/bin/%s', ngx.var.tuf_repo_url, dst_url, ngx.var.version, ngx.var.os, ngx.var.arch, ngx.var.target), 302)
else
  ngx.log(ngx.WARN, string.format('Got status %s when trying to request version (URL - %s)', res.status, version_request_url))
  ngx.status = res.status
  return ngx.exit(res.status)
end
