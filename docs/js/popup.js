$(document).ready(function() {
  $('[data-open-popup]').on('click', function(e) {
    e.preventDefault();
    $('[data-popup="' + $(this).data('open-popup') + '"]').addClass('active');
  });
  $('[data-close-popup]').on('click', function(e) {
    e.preventDefault();
    $('[data-popup]').removeClass('active');
  });
  $(document).on('keyup',function(evt) {
    if (evt.keyCode === 27) {
      $('[data-popup]').removeClass('active');
    }
  });
  $('[data-popup] .popup__content').click(function(e) {
    e.stopPropagation();
  });
  $('[data-popup]').click(function() {
    $('[data-popup]').removeClass('active');
  });
});
