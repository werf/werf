$(document).ready(function() {
  if ($('#introduction-presentation').length) {
    var ip_slides = [];
    var ip_magic = new ScrollMagic.Controller();
    var ip_offset = ($('#introduction-presentation').position().top - window.innerHeight/2) + 400;
    var ip_offset_step = 400;

    var zi = 1000;
    $('.introduction-presentation__slide').each(function() {
      $(this).css('z-index', zi); zi--;
      ip_slides.push($(this)[0]);
    })

    function updateControls(slide_index) {
      var target = $(`.introduction-presentation__controls-selector a[data-presentation-selector-option=${slide_index}]`);

      $(`.introduction-presentation__controls-selector a`).removeClass('active');
      $(target).addClass('active');

      $('.introduction-presentation__controls-step').html($(target).html());
      $('.introduction-presentation__controls-stage').html($(target).data('presentation-selector-stage'));
    }

    function addIpScene(slide, slide_index) {
      new ScrollMagic.Scene({
        duration: 100,
        offset: ip_offset
      })
      .setTween(
        new TimelineMax().to(slide, {
          opacity: '0',
          display: 'none'
        }, 0)
      )
      .on('start', (event) => {
        if (event.scrollDirection == 'FORWARD') {
          updateControls(slide_index+1)
        } else {
          updateControls(slide_index)
        }
      })
      .addTo(ip_magic);

      ip_offset = ip_offset + ip_offset_step;
    }
    new ScrollMagic.Scene({
      duration: (ip_slides.length*ip_offset_step) + ip_offset_step/2,
      triggerElement: '#introduction-presentation',
      offset: window.innerHeight*0.7/2
    })
    .setPin('#introduction-presentation')
    .addTo(ip_magic);

    // Hide controls
    new ScrollMagic.Scene({
      duration: 100,
      offset: (ip_slides.length*ip_offset_step) + ip_offset_step/2
    })
    .setTween(
      new TimelineMax().to('#introduction-presentation-controls', {
        opacity: '0',
      }, 0)
    )
    .addTo(ip_magic);

    ip_slides.forEach(function(ip_slide, index) {
      addIpScene(ip_slide, index);
    })

    $('.introduction-presentation__controls-nav').on('click', (e) => {
      e.preventDefault();
      $('.introduction-presentation__controls-selector').toggle();
    })

    $('.introduction-presentation__controls-selector a').on('click', function(e) {
      e.preventDefault();
      var slide = $(this).data('presentation-selector-option');
      var scroll = ($('#introduction-presentation').position().top - window.innerHeight/2) + 400*(slide+2);

      window.scrollTo(0, scroll);
      $('.introduction-presentation__controls-selector').toggle();
    })
  }
});