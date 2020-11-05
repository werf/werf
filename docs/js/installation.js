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
    channel: 'alpha',
    os: default_os,
    method: 'multiwerf'
  }

  function doInstallSelect(group, param) {
    $(`[data-install-tab-group="${group}"]`).removeClass('active');
    $(`[data-install-tab="${param}"]`).addClass('active');

    $(`[data-install-content-group="${group}"]`).removeClass('active');
    $(`[data-install-content="${param}"]`).addClass('active');

    $(`[data-install-info="${group}"]`).text($(`[data-install-tab="${param}"]`).text());
  }

  function installSelect(group, param) {
    if (group == "version" && param == "1.2") {
        $(`[data-install-tab="rock-solid"]`).hide();
        $(`[data-install-tab="stable"]`).hide();
        $(`[data-install-tab="ea"]`).hide();

        doInstallSelect(group, param)
        doInstallSelect("channel", "beta")
        return
    } else if (group == "version") {
        $(`[data-install-tab="rock-solid"]`).show();
        $(`[data-install-tab="stable"]`).show();
        $(`[data-install-tab="ea"]`).show();
        $(`[data-install-tab="beta"]`).show();
    }

    doInstallSelect(group, param)
  }

  $('[data-install-tab]').on('click', function() {
    installSelect($(this).data('install-tab-group'), $(this).data('install-tab'));
  })

  Object.keys(defaults).forEach(function(key) {
    installSelect(key, defaults[key]);
  });
})
