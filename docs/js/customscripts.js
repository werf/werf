
$('#mysidebar').height($(".nav").height());


$( document ).ready(function() {
    var wh = $(window).height();
    var sh = $("#mysidebar").height();
    
    if (sh + 100 > wh) {
        $( "#mysidebar" ).parent().addClass("layout-sidebar__sidebar_a");
    }
    // activate tooltips. although this is a bootstrap js function, it must be activated this way in your theme.
    $('[data-toggle="tooltip"]').tooltip({
        placement : 'top'
    });

    /**
     * AnchorJS
     */
    anchors.add('h2,h3,h4,h5');

});

// needed for nav tabs on pages. See Formatting > Nav tabs for more details.
// script from http://stackoverflow.com/questions/10523433/how-do-i-keep-the-current-tab-active-with-twitter-bootstrap-after-a-page-reload
$(function() {
    var json, tabsState;
    $('a[data-toggle="pill"], a[data-toggle="tab"]').on('shown.bs.tab', function(e) {
        var href, json, parentId, tabsState;

        tabsState = localStorage.getItem("tabs-state");
        json = JSON.parse(tabsState || "{}");
        parentId = $(e.target).parents("ul.nav.nav-pills, ul.nav.nav-tabs").attr("id");
        href = $(e.target).attr('href');
        json[parentId] = href;

        return localStorage.setItem("tabs-state", JSON.stringify(json));
    });

    tabsState = localStorage.getItem("tabs-state");
    json = JSON.parse(tabsState || "{}");

    $.each(json, function(containerId, href) {
        return $("#" + containerId + " a[href=" + href + "]").tab('show');
    });

    $("ul.nav.nav-pills, ul.nav.nav-tabs").each(function() {
        var $this = $(this);
        if (!json[$this.attr("id")]) {
            return $this.find("a[data-toggle=tab]:first, a[data-toggle=pill]:first").tab("show");
        }
    });
});

// Load versions and append them to topnavbar
$( document ).ready(function() {
    $.getJSON('/assets/channels.json').success(function (resp) {
    var releasesInfo = resp;

    var match = document.location.href.match(/(v[^/]*|master|latest)/);
    var currentRelease, currentChannel;
    if (match) {
      currentRelease = match[1];
      currentChannel = releasesInfo.releases[currentRelease] && releasesInfo.releases[currentRelease][0];
        if ((currentRelease == 'master') || (currentRelease == 'latest')) {
            currentChannel = currentRelease;
        }
    }

        console.log('releasesInfo: ', releasesInfo);
        console.log('currentRelease: ', currentRelease);
        console.log('currentChannel: ', currentChannel);


    if (!( currentRelease in releasesInfo['releases'] )) {
      $('#outdatedWarning').addClass('active');
    }
        var menu = $('#doc-versions-menu');
    menu.addClass('header__menu-item header__menu-item_parent');
    var toggler = $('<a href="#">');
    // toggler.addClass('dropdown-toggle');
    toggler.append(currentChannel || 'Versions');
    // toggler.append($('<b class="caret">'));

    menu.html(toggler);
    var submenu = $('<ul class="header__submenu">');
    $.each(releasesInfo.orderedReleases, function(i, release) {
      var link = $('<a href="/' + release + '">');
      link.append(release);
      if (releasesInfo.releases[release]) {
        $.each(releasesInfo.releases[release], function(j, channel){
          link.append('&nbsp; / ');
          link.append(channel);
        });
      }
        var item = $('<li class="header__submenu-item">');
      // if (release == currentRelease)
      //   item.addClass('dropdownActive');
      item.html(link);
      submenu.append(item);
    });
    menu.append(submenu);
  });

  // Update github counters 
  $( document ).ready(function() {
    $.get("https://api.github.com/repos/flant/werf", function(data) {
      $(".gh_counter").each(function( index ) {
        $(this).text(data.stargazers_count)
      });
    });
  });

  $( document ).ready(function() {
    var $header = $('.header');
    function updateHeader() {
      if ($(document).scrollTop() == 0) {
        $header.removeClass('header_active');
      } else {
        $header.addClass('header_active');
      }
    }
    $(window).scroll(function() {
      updateHeader();
    });
    updateHeader();
  });

  $( document ).ready(function() {
    $('.header__menu-icon_search').on('click tap', function() {
      $('.topsearch').toggleClass('topsearch_active');
      $('.header').toggleClass('header_search');
      if ($('.topsearch').hasClass('topsearch_active')) {
        $('.topsearch__input').focus();
      } else {
        $('.topsearch__input').blur();
      }
    });

    $('body').on('click tap', function(e) {
      if ($(e.target).closest('.topsearch').length === 0 && $(e.target).closest('.header').length === 0) {
        $('.header').toggleClass('header_search');
        $('.topsearch').removeClass('topsearch_active');
      }
    });
  });

});
