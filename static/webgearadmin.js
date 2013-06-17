$(function() {
    function updateStatus() {
        var section = $(this),
            statusTable = section.find('table[name=statuses]'),
            address = section.attr('id');
        $.get("/api/" + address + "/workers")
        .done(function(data) {
            statusTable.find('tbody tr').remove();
            $.each(data, function(key, value) {
                statusTable.find('tbody').append("<tr>" +
                    "<td>" + value.FunctionName + "</td>" +
                    "<td>" + value.JobTotal + "</td>" +
                    "<td>" + value.JobRunning + "</td>" +
                    "<td>" + value.WorkerCount + "</td>" +
                "</tr>");
            });
        })
        .fail(function() {
            clearInterval(section.data('intervalName'));
        });
    }

    $('section').each(function() {
        var intervalName = setInterval(updateStatus.bind(this), 1000);
        $(this).data('intervalName', intervalName)
        updateStatus.call(this);
    });
});
