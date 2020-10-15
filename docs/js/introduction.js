$(document).ready(function() {
  if ($('#introduction-presentation')) {
    var ip_slides = ['#ip_1', '#ip_2', '#ip_3'];
    var ip_magic = new ScrollMagic.Controller();
    var ip_offset = 0;
    var ip_offset_step = 400;

    function addIpScene(slide) {
      new ScrollMagic.Scene({duration: 100, offset: ip_offset})
      .setTween(new TimelineMax().to(slide, {opacity: '0'}, 0)).addTo(ip_magic);
      ip_offset = ip_offset + ip_offset_step;
    }

    // Pin scheme
    new ScrollMagic.Scene({duration: ip_slides.length*ip_offset_step, offset: -1})
    .setPin('#introduction-presentation')
    .addTo(ip_magic);

    ip_slides.forEach(function(ip_slide) {
      addIpScene(ip_slide);
    })
  }
});