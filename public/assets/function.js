$(function(){
  /* Problems quick navigate */
  $("#quick-problem").submit(function(e) {
    e.preventDefault();
    var pid;
    pid = $('#quick-problem-id').val();
    if (pid === "") {
      return;
    }
    return window.location.href = "/problems/" + pid;
  });
});
