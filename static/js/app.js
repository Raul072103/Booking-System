function Prompt() {
    let toast = function (c) {
        const {
            msg = "",
            icon = "success",
            position = "top-end",
        } = c;
        const Toast = Swal.mixin({
            toast: true,
            title: msg,
            icon: icon,
            position: position,
            showConfirmButton: false,
            timer: 3000,
            timerProgressBar: true,
            didOpen: (toast) => {
                toast.onmouseenter = Swal.stopTimer;
                toast.onmouseleave = Swal.resumeTimer;
            }
        });

        Toast.fire({});
    }

    let success = function (c) {
        const {
            msg = "",
            title = "",
            footer = "",
        } = c;
        Swal.fire({
            icon: "success",
            title: title,
            text: msg,
            footer: footer
        });
    }

    let error = function (c) {
        const {
            msg = "",
            title = "",
            footer = "",
        } = c;
        Swal.fire({
            icon: "error",
            title: title,
            text: msg,
            footer: footer
        });
    }

    let custom = async function (c) {
        const {
            icon = "",
            msg = "",
            title = "",
            showConfirmButton = true,
        } = c;

        const { value: result } = await Swal.fire({
            icon: icon,
            title: title,
            html: msg,
            backdrop: false,
            focusConfirm: false,
            showCancelButton: true,
            showConfirmButton: showConfirmButton,
            willOpen: () => {
                if (c.willOpen !== undefined)
                    c.willOpen()
            },
            preConfirm: () => {
                if (c.preConfirm !== undefined)
                    c.preConfirm()
            },
            didOpen: () => {
                if (c.didOpen !== undefined)
                    c.didOpen()
            }
        })

        if (result) {
            if (result.dismiss !== Swal.DismissReason.cancel) {
                if (result.value !== "") {
                    if (c.callback !== undefined) {
                        c.callback(result);
                    }
                } else {
                    c.callback(false);
                }
            } else {
                c.callback(false);
            }
        }
    }

    return {
        toast: toast,
        success: success,
        error: error,
        custom: custom,
    }
}

function PromptDateSelectionForRoom(roomId, document, csrfToken) {
    let html = `
			<form id="check-availability-form" action="" method="post" novalidate class="needs-validation">
				<div class="row">
					<div class="col">
						<div class="row" id="reservation-dates-modal">
							<div class="col">
								<input required class="form-control" disabled type="text" name="start" id="start" placeholder="Arrival">
							</div>
							<div class="col">
								<input required class="form-control" disabled type="text" name="end" id="end" placeholder="Departure">
							</div>
						</div>
					</div>
				</div>
			</form>
			`
        let attention = Prompt()
        return attention.custom({
            msg: html,
            title: "Choose your dates",
            willOpen: () => {
                const elem = document.getElementById("reservation-dates-modal");
                const rp = new DateRangePicker(elem, {
                    format: 'yyyy-mm-dd',
                    showOnFocus: true,
                    minDate: new Date(),
                })
            },
            didOpen: () => {
                document.getElementById("start").removeAttribute("disabled")
                document.getElementById("end").removeAttribute("disabled")
            },
            preConfirm: () => {
                return [
                    document.getElementById("start").value,
                    document.getElementById("end").value
                ];
            },
            callback: function(result) {
                console.log("called")

                let form = document.getElementById("check-availability-form")
                let formData = new FormData(form);
                formData.append("csrf_token", csrfToken);
                formData.append("room_id", roomId);

                fetch('/search-availability-json', {
                    method: "post",
                    body: formData,
                })
                    .then(response => response.json())
                    .then(data => {
                        if (data.ok) {
                            attention.custom({
                                icon: 'success',
                                showConfirmButton: false,
                                msg: '<p>Room is available</p>'
                                    + '<p><a href="/book-room?id='
                                    + data.room_id
                                    + '&s='
                                    + data.start_date
                                    + '&e='
                                    + data.end_date
                                    + '" class="btn btn-primary">'
                                    + 'Book now!</a></p>'
                            })
                        } else {
                            attention.error({
                                msg: "No availability",
                            })
                        }
                    })
            }
        })
}