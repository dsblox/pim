function pimShowError(message) {
	$('#err').html("<div class=\"alert alert-danger alert-dismissible show\" role=\"alert\" id=errortext>"
            + message + 
            "<button type=\"button\" class=\"close\" aria-label=\"Close\" onclick=\"$('#err').hide()\"> \
              <span aria-hidden=\"true\">&times;</span> \
            </button> \
          </div>");
      $('#err').show();
 }

 function pimAjaxError(response) {
 	r = JSON.parse(response);
 	pimShowError(r.msg)
 }