$(document).ready(function() {
  var default_os;
  if (bowser.windows)
    default_os = 'windows'
  else if (bowser.mac)
    default_os = 'macos'
  else
    default_os = 'linux'

  var defaults = {
    version: '1.2',
    channel: 'stable',
    os: default_os,
    method: 'multiwerf'
  }

  function installSelect(group, param) {
    $(`[data-install-tab-group="${group}"]`).removeClass('active');
    $(`[data-install-tab="${param}"]`).addClass('active');

    $(`[data-install-content-group="${group}"]`).removeClass('active');
    $(`[data-install-content="${param}"]`).addClass('active');

    $(`[data-install-info="${group}"]`).text($(`[data-install-tab="${param}"]`).text());
  }

  $('[data-install-tab]').on('click', function() {
    installSelect($(this).data('install-tab-group'), $(this).data('install-tab'));
  })

  Object.keys(defaults).forEach(function(key) {
    installSelect(key, defaults[key]);
  });
})
