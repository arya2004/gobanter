let socket = null;
    
$(document).ready(function () {
    const offline = `<span class="badge bg-danger">Not connected</span>`;
    const online = `<span class="badge bg-success">Connected</span>`;

    const $statusDiv = $("#status");
    const $output = $("#output");
    const $userField = $("#username");
    const $messageField = $("#message");
    const $onlineUsers = $("#online_users");

    // Reconnecting WebSocket Initialization
    socket = new ReconnectingWebSocket("ws://localhost:8080/ws", null, { debug: true, reconnectInterval: 3000 });

    // WebSocket Events
    socket.onopen = function () {
        console.log("connected!!");
        $statusDiv.html(online);
    };

    socket.onclose = function () {
        console.log("connection closed!");
        $statusDiv.html(offline);
    };

    socket.onerror = function () {
        console.log("there was an error");
        $statusDiv.html(offline);
    };

    socket.onmessage = function (msg) {
        const data = JSON.parse(msg.data);
        console.log("Action:", data.action);

        switch (data.action) {
            case "list_users":
                $onlineUsers.empty();
                if (data.connected_users.length > 0) {
                    $.each(data.connected_users, function (index, user) {
                        $onlineUsers.append(`<li class="list-group-item">${user}</li>`);
                    });
                }
                break;

            case "broadcast":
                $output.append(`<div>${data.message}</div>`);
                $output.scrollTop($output.prop("scrollHeight"));
                break;
        }
    };

    // Username Field Change
    $userField.on("change", function () {
        const jsonData = {
            action: "username",
            username: $(this).val()
        };
        console.log(jsonData);
        socket.send(JSON.stringify(jsonData));
    });

    // Message Field Enter Key
    $messageField.on("keydown", function (event) {
        if (event.key === "Enter") {
            if (!socket || socket.readyState !== WebSocket.OPEN) {
                alert("No connection");
                return false;
            }

            if (!$userField.val() || !$messageField.val()) {
                alert("Enter username and message!");
                return false;
            } else {
                sendMessage();
            }

            event.preventDefault();
        }
    });

    // Send Button Click
    $("#sendButton").on("click", function () {
        if (!$userField.val() || !$messageField.val()) {
            alert("Enter username and message!");
            return false;
        } else {
            sendMessage();
        }
    });

    // WebSocket Disconnect on Page Unload
    $(window).on("beforeunload", function () {
        console.log("leaving ;(");
        if (socket && socket.readyState === WebSocket.OPEN) {
            const jsonData = { action: "left" };
            socket.send(JSON.stringify(jsonData));
        }
    });

    // Function to Send Message
    function sendMessage() {
        const jsonData = {
            action: "broadcast",
            username: $userField.val(),
            message: $messageField.val()
        };
        socket.send(JSON.stringify(jsonData));
        $messageField.val("");
    }
});