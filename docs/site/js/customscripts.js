
$('#mysidebar').height($(".nav").height());


document.addEventListener("DOMContentLoaded", function() {
  /**
   * AnchorJS
   */
  anchors.add('h2,h3,h4,h5');

  var referenceAnchors;
  if ($('.line-number-anchor')) {
      referenceAnchors = new AnchorJS();
      referenceAnchors.options = {
          placement: 'left'
      };
      referenceAnchors.add('.line-number-anchor');
  }
});

$( document ).ready(function() {
  $('.header__menu').addClass('header__menu_active');
});

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

// Update GitHub stats
$(document).ready(function () {
  var github_requests = [],
  github_stats = JSON.parse(localStorage.getItem('werf_github_stats')) || null;

  function getGithubRequests() {
    $('[data-roadmap-step]').each(function () {
      var $step = $(this);
      github_requests.push($.get('https://api.github.com/repos/werf/werf/issues/' + $step.data('roadmap-step'), function (data) {
        github_stats['issues'][$step.data('roadmap-step')] = (data.state === 'closed');
      }));
    });
    github_requests.push($.get("https://api.github.com/repos/werf/werf", function (data) {
      github_stats['stargazers'] = data.stargazers_count
    }));
    return github_requests;
  }

  function updateGithubStats() {
    $('.gh_counter').each(function () {
      $(this).text(github_stats['stargazers']);
    });
    $('[data-roadmap-step]').each(function () {
      var $step = $(this);
      if (github_stats['issues'][$step.data('roadmap-step')] == true) $step.addClass('roadmap__steps-list-item_closed');
    });
  }

  if (github_stats == null || Date.now() > (github_stats['updated_on'] + 1000 * 60 * 60)) {
    github_stats = {'updated_on': Date.now(), 'issues': {}, 'stargazers': 0};
    $.when.apply($, getGithubRequests()).done(function() {
      updateGithubStats();
      localStorage.setItem('werf_github_stats', JSON.stringify(github_stats));
    });
  } else {
    updateGithubStats();
  }
});

$(document).ready(function () {
  var $header = $('.header');

  function updateHeader() {
    if ($(document).scrollTop() == 0) {
      $header.removeClass('header_active');
    } else {
      $header.addClass('header_active');
    }
  }

  $(window).scroll(function () {
    updateHeader();
  });
  updateHeader();
});

$(document).ready(function () {
  $('.header__menu-icon_search').on('click tap', function () {
    $('.topsearch').toggleClass('topsearch_active');
    $('.header').toggleClass('header_search');
    if ($('.topsearch').hasClass('topsearch_active')) {
      $('.topsearch__input').focus();
    } else {
      $('.topsearch__input').blur();
    }
  });

  $('body').on('click tap', function (e) {
    if ($(e.target).closest('.topsearch').length === 0 && $(e.target).closest('.header').length === 0) {
      $('.header').removeClass('header_search');
      $('.topsearch').removeClass('topsearch_active');
    }
  });

  let CurrentGroup = getDocGroupFromURL();
  $('input[name="sitesearch"]').attr("value",function(index,currentvalue){
    if (CurrentGroup) {
        return currentvalue + "/documentation/" + CurrentGroup + "/"
    } else {
        return currentvalue
    }
  })

});

$(document).ready(function() {
  var adjustAnchor = function() {
      var $anchor = $(':target'), fixedElementHeight = 120;
      if ($anchor.length > 0) {
        $('html, body').stop().animate({
          scrollTop: $anchor.offset().top - fixedElementHeight
        }, 200);
      }
  };
  $(window).on('hashchange load', function() {
      adjustAnchor();
  });
});

$(document).ready(function(){
  // wait until fonts are loaded
  setTimeout(function() {
    $('.publications__list').masonry({
      itemSelector: '.publications__post',
      columnWidth: '.publications__sizer'
    })
  }, 500)
});

$(document).ready(function(){

  $('h1:contains("Installation")').each(function( index ) {
    var $title = $(this);
    var $btn1 = $title.next('p');
    var $btn2 = $btn1.next('p');
    var $btn3 = $btn2.next('p');

    var new_btns = $('<div class="publications__install-btns">');
    new_btns.append($($btn1.html()).addClass('releases__btn'));
    new_btns.append($($btn2.html()).addClass('releases__btn'));
    new_btns.append($($btn3.html()).addClass('releases__btn'));

    $btn1.remove();
    $btn2.remove();
    $btn3.remove();
    $title.after(new_btns);
  });
});

// Presentation

$(document).ready(function() {
  if($('#presentation').length) {
    var magic = new ScrollMagic.Controller();

    // Pin intro cheme
    var intro_scheme_duration = window.innerHeight;
    intro_scheme_duration = intro_scheme_duration > 1080 ? 1080 : intro_scheme_duration;
    intro_scheme_duration = intro_scheme_duration < 590 ? 590 : intro_scheme_duration;
    new ScrollMagic.Scene({duration: intro_scheme_duration, offset: -10})
    .setPin('#intro-scheme', {pushFollowers: false})
    .addTo(magic);

    // Pin bg
    new ScrollMagic.Scene({duration: 4000, offset: -10})
    .setPin('#intro-bg', {pushFollowers: false})
    .addTo(magic);

    // Pin scheme
    new ScrollMagic.Scene({triggerElement: '#presentation', triggerHook: 0, duration: 4000, offset: -90})
    .setPin('#presentation', {pushFollowers: false})
    .addTo(magic);

    // Pin notes
    new ScrollMagic.Scene({duration: 400, triggerElement: '#presentation-notes-1',  offset: 100})
    .setPin('#presentation-notes-1', {pushFollowers: false})
    .addTo(magic);
    new ScrollMagic.Scene({duration: 400, triggerElement: '#presentation-notes-2', offset: 100})
    .setPin('#presentation-notes-2', {pushFollowers: false})
    .addTo(magic);
    new ScrollMagic.Scene({duration: 1000, triggerElement: '#presentation-notes-3', offset: 100})
    .setPin('#presentation-notes-3', {pushFollowers: false})
    .addTo(magic);

    // Move away title
    new ScrollMagic.Scene({triggerElement: '#presentation', triggerHook: 0, duration: 250, offset: 25}).setTween(
      new TimelineMax()
      .to('#presentation-title', {x: '-2000px', opacity: 0}, 0)
      .to('#intro-bg', {x: '-1500px'}, 0)
    )
    .addTo(magic);

    // Hide arrows & smart
    new ScrollMagic.Scene({triggerElement: '#presentation', triggerHook: 0, duration: 100, offset: 400}).setTween(
      new TimelineMax()
      .to('#scheme_git', {opacity: '0.2'}, 0)
      .to('#scheme_docker_registry', {opacity: '0.2'}, 0)
      .to('#scheme_k8s', {opacity: '0.2'}, 0)
      .to('#scheme_werf', {opacity: '0.2'}, 0)
      .to('#scheme_arrows_gw', {opacity: '0'}, 0)
      .to('#scheme_arrows_wd', {opacity: '0'}, 0)
      .to('#scheme_arrows_wk', {opacity: '0'}, 0)
      .to('#scheme_smart_2', {opacity: '0'}, 0)
      .to('#scheme_smart', {opacity: '0'}, 0)
    ).addTo(magic);

    // Git
    new ScrollMagic.Scene({triggerElement: '#presentation', triggerHook: 0, duration: 200, offset: 600}).setTween(
      new TimelineMax()
      .to('#scheme_git', {opacity: '1'}, 0)
    ).addTo(magic);

    // Git -> Werf, show
    new ScrollMagic.Scene({triggerElement: '#presentation', triggerHook: 0, duration: 200, offset: 1000}).setTween(
      new TimelineMax()
      .to('#scheme_git', {opacity: '1'}, 0)
      .to('#scheme_arrows_gw', {opacity: '1'}, 0)
      .to('#scheme_werf', {opacity: '1'}, 0)
    ).addTo(magic);

    // Werf -> Docker Registry, show
    new ScrollMagic.Scene({triggerElement: '#presentation', triggerHook: 0, duration: 200, offset: 1400}).setTween(
      new TimelineMax()
      .to('#scheme_git', {opacity: '0.2'}, 0)
      .to('#scheme_docker_registry', {opacity: '1'}, 0)
      .to('#scheme_arrows_gw', {opacity: '0.2'}, 0)
      .to('#scheme_arrows_wd', {opacity: '1'}, 0)
      .to('#scheme_smart_2', {opacity: '1'}, 0)
    ).addTo(magic);

    // Werf -> Docker Registry, sync
    new ScrollMagic.Scene({triggerElement: '#presentation', triggerHook: 0, duration: 500, offset: 1800}).setTween(
      TweenMax
      .fromTo('#scheme_smart_icon_update_arrows_2', 1,
      {rotation: '0'}, {rotation: '-720', transformOrigin: '50% 50%', repeat: -1, ease: "power1.out"})
      .duration(2)
    ).addTo(magic);

    // Werf -> Docker Registry, show info
    new ScrollMagic.Scene({triggerElement: '#presentation', triggerHook: 0, duration: 200, offset: 2200}).setTween(
      new TimelineMax()
      .to('#scheme_smart_icon_update_2', {opacity: '0'})
      .to('#scheme_smart_icon_check_2', {opacity: '1'})
    ).addTo(magic);

    // Werf -> Kubernetes, show
    new ScrollMagic.Scene({triggerElement: '#presentation', triggerHook: 0, duration: 200, offset: 2600}).setTween(
      new TimelineMax()
      .to('#scheme_docker_registry', {opacity: '0.2'}, 0)
      .to('#scheme_arrows_wd', {opacity: '0.2'}, 0)
      .to('#scheme_arrows_wk', {opacity: '1'}, 0)
      .to('#scheme_smart', {opacity: '1'}, 0)
      .to('#scheme_k8s', {opacity: '1'}, 0)
    ).addTo(magic);

    // Werf -> Kubernetes, sync
    new ScrollMagic.Scene({triggerElement: '#presentation', triggerHook: 0, duration: 500, offset: 3000}).setTween(
      TweenMax
      .fromTo('#scheme_smart_icon_update_arrows', 1,
      {rotation: '0'}, {rotation: '-720', transformOrigin: '50% 50%', repeat: -1, ease: "power1.out"})
      .duration(2)
    ).addTo(magic);

    // Werf -> Docker Registry, show info
    new ScrollMagic.Scene({triggerElement: '#presentation', triggerHook: 0, duration: 200, offset: 3400}).setTween(
      new TimelineMax()
      .to('#scheme_smart_icon_update', {opacity: '0'})
      .to('#scheme_smart_icon_check', {opacity: '1'})
    ).addTo(magic);

    // Full
    new ScrollMagic.Scene({triggerElement: '#presentation', triggerHook: 0, duration: 200, offset: 3800}).setTween(
      new TimelineMax()
      .to('#scheme_smart_icon_update', {opacity: '0'})
      .to('#scheme_smart_icon_check', {opacity: '1'})
      .to('#scheme_docker_registry', {opacity: '1'}, 0)
      .to('#scheme_arrows_wd', {opacity: '1'}, 0)
      .to('#scheme_git', {opacity: '1'}, 0)
      .to('#scheme_arrows_gw', {opacity: '1'}, 0)
    ).addTo(magic);
  }
});


// Intro scheme

$(document).ready(function() {
  if($('#intro-animation').length) {
    var ia = {};
    ia.dot =          '#intro-animation-dot';
    ia.dot2 =         '#intro-animation-dot2';
    ia.k8s =          '#intro-animation-k8s';
    ia.git =          '#intro-animation-git';
    ia.docker =       '#intro-animation-docker';
    ia.helm =         '#intro-animation-helm';
    ia.werf =         '#intro-animation-werf';
    ia.werf_k8s =     '#intro-animation-werf-k8s';
    ia.werf_git =     '#intro-animation-werf-git';
    ia.werf_docker =  '#intro-animation-werf-docker';
    ia.werf_helm =    '#intro-animation-werf-helm';

    function moveDot(data) {
      data.timeline
      .to(data.dot,  {duration: 2, ease: 'power1.inOut',
                            motionPath: {
                              path: data.path,
                              align: data.path,
                              alignOrigin: [0.5, 0.5],
                              autoRotate: false,
                              start: 0,
	                            end: data.reverse ? -1 : 1,
                            }
                          }, data.delay)
      .to(data.dot,  {opacity: 1, duration: 0.25}, '-=2')
      .to(data.dot,  {opacity: 0, duration: 0.25}, '-=0.25');
    }

    var tl = gsap.timeline({repeat: -1, repeatDelay: 1, delay: 1});
    // Dot git werf
    moveDot({timeline: tl,  dot: ia.dot,    path: ia.werf_git,      reverse: false,     delay: '=0'   });
    // Dot werf docker
    moveDot({timeline: tl,  dot: ia.dot,    path: ia.werf_docker,   reverse: false,     delay: '=0'   });
    // Dot werf helm
    moveDot({timeline: tl,  dot: ia.dot2,   path: ia.werf_helm,     reverse: true,      delay: '-=2'  });
    // Dot docker werf
    moveDot({timeline: tl,  dot: ia.dot,    path: ia.werf_docker,   reverse: true,      delay: '=0'   });
    // Dot helm werf
    moveDot({timeline: tl,  dot: ia.dot2,   path: ia.werf_helm,     reverse: false,     delay: '-=2'  });
    // Dot werf k8s
    moveDot({timeline: tl,  dot: ia.dot,    path: ia.werf_k8s,      reverse: true,      delay: '=0'   });
  }
});

// News and updates

$(document).ready(function () {
  var $nau = $('.news-and-updates');

  function updateNauVisibility() {
    if ($(document).scrollTop() < 50) {
      $nau.addClass('news-and-updates__active');
    } else {
      $nau.removeClass('news-and-updates__active');
    }
  }

  function updateNauContent() {
    var $releases_container = $('#nau-releases');
    nau_data.releases.slice(0, 5).forEach(function(item) {
      $releases_container.append(
      `<div class="news-and-updates__item">
          <a href="${item.html_url}" target="_blank" class="news-and-updates__item-link">
              ${item.tag_name}
          </a>
          <div class="news-and-updates__item-sub">
              ${moment(item.published_at).format('DD.MM.YYYY') }
          </div>
      </div>`);
    });

    var $news_container = $('#nau-news');
    $news_container.append(
    `<a href="${nau_data.news.url}" target="_blank" class="news-and-updates__item-link">
    ${nau_data.news.title}
    </a>
    <div class="news-and-updates__item-sub">
      ${moment(nau_data.news.date).format('DD.MM.YYYY') }
    </div>`);
  }

  if ($nau) {
    var nau_data = JSON.parse(localStorage.getItem('werf_news')) || null;
    var nau_requests = [
      $.get("https://api.github.com/repos/werf/werf/releases", function (data) {
        nau_data.releases = data;
      }),
      $.get("https://zapier.com/engine/rss/9718388/werf-io-tweets", function (data) {
        var rss = data;
        var rss_items = rss.querySelectorAll('item');
        var rss_result = null;

        var i = 0;
        do {
          var rss_item = rss_items[i];
          var rss_item_description = rss_item.querySelector('description');
          if (rss_item_description) {
            if (rss_item_description.innerHTML.indexOf('#changelog')) {
              rss_result = rss_item;
            }
          }
          i = i + 1;
        } while (i < rss_items.length && rss_result == null);
        if (rss_result == null) { rss_result = rss_items[0] }

        nau_data.news = {
          title: rss_result.querySelector('description').innerHTML,
          date: rss_result.querySelector('pubDate').innerHTML,
          url: rss_result.querySelector('title').innerHTML
        };
      })
    ];

    if (nau_data == null || Date.now() > (nau_data['updated_on'] + 1000 * 60 * 60)) {
      nau_data = {'updated_on': Date.now(), 'news': {}, 'releases': []};
      $.when.apply($, nau_requests).done(function() {
        updateNauContent();
        localStorage.setItem('werf_news', JSON.stringify(nau_data));
      });
    } else {
      updateNauContent();
    }

    var throttled = _.throttle(updateNauVisibility, 100);
    $(window).scroll(throttled);
    updateNauVisibility();
  }
});

function getDocGroupFromURL() {
  let result = window.location.pathname.match(/^\/documentation\/(v\d+\.\d+)/);
  if ( result && result[1] ) {
    return result[1];
  }
  return null;
}

function getDocVersionFromURL() {
  let result = window.location.pathname.match(/\/documentation\/(v\d+\.\d+([^/]+)?)\/.*/);
  if ( result ) {
    if  ( result.length > 2 && result[2] ) {
      return result[1].replace('-plus-','+');
    } else {
      return "master";
    }
  }
  return null;
}

function getDocVersionFromPage() {
  let result = $("#versionNumber").text();
  if ( !result) {
    result = "";
  }
  return result;
}

function checkURLExist(url, callback) {

    if ( ! $.isFunction(callback)) {
       throw Error('Not a valid callback');
    }

    $.ajax({
        type: 'HEAD',
        url: url,
        success: $.proxy(callback, this, true),
        error: $.proxy(callback, this, false)
    });
}
