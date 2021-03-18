$( document ).ready(function() {
  $('.details__summary').on('click tap', function() {
    $(this).closest('.details').toggleClass('active');
  });
});