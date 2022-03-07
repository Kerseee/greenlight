document.addEventListener('DOMContentLoaded', function(){
    fetch("http://localhost:8080/v1/healthcheck").then(
        function(resp) {
            resp.text().then(
                function(text) {
                    document.getElementById("output").innerHTML = text;
                }
            );
        },
        function(err) {
            document.getElementById("output").innerHTML = err
        }
    )
});