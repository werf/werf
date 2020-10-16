$(document).ready(function() {
  if ($('#introduction-presentation')) {
    var ip_slides = [];
    var ip_magic = new ScrollMagic.Controller();
    var ip_offset = ($('#introduction-presentation').position().top - window.innerHeight/2) + 300;
    var ip_offset_step = 400;

    var zi = 1000;
    $('.introduction-presentation__slide').each(function() {
      $(this).css('z-index', zi); zi--;
      ip_slides.push($(this)[0]);
    })

    function addIpScene(slide) {
      new ScrollMagic.Scene({
        duration: 100,
        offset: ip_offset
      })
      .setTween(
        new TimelineMax().to(slide, {
          opacity: '0',
          display: 'none'
        }, 0)
      ).addTo(ip_magic);
      ip_offset = ip_offset + ip_offset_step;
    }

    // Pin scheme
    new ScrollMagic.Scene({
      duration: (ip_slides.length*ip_offset_step) + ip_offset_step/2,
      triggerElement: '#introduction-presentation',
      offset: 150
    })
    .setPin('#introduction-presentation')
    .addTo(ip_magic);

    ip_slides.forEach(function(ip_slide) {
      addIpScene(ip_slide);
    })
  }
});