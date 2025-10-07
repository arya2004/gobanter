let socket = null;
let typingTimer;

$(document).ready(function () {
    const offline = `<span class="badge bg-danger">Not connected</span>`;
    const online = `<span class="badge bg-success">Connected</span>`;

    const $statusDiv = $("#status");
    const $output = $("#output");
    const $userField = $("#username");
    const $messageField = $("#message");
    const $onlineUsers = $("#online_users");
    const $recipientSelect = $("#recipient-select"); //  fixed missing #

    socket = new ReconnectingWebSocket("ws://localhost:8080/ws", null, {
        debug: true,
        reconnectInterval: 3000,
    });

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
                updateRecipientList(data.connected_users);
                break;

            case "typing":
                showTypingIndicator(data.from);
                break;

            case "broadcast":
                displayBroadcastMessage(data);
                break;

            case "private":
                displayPrivateMessage(data);
                break;

            case "error":
                displayError(data.message);
                break;
        }
    };

    $userField.on("change", function () {
        const jsonData = {
            action: "username",
            username: $(this).val(),
        };
        console.log(jsonData);
        socket.send(JSON.stringify(jsonData));
    });

    $messageField.on("input", function () {
        if (typingTimer) clearTimeout(typingTimer);
        const jsonData = {
            action: "typing",
            username: $userField.val(),
        };
        socket.send(JSON.stringify(jsonData));
    });

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

    $("#sendButton").on("click", function () {
        if (!$userField.val() || !$messageField.val()) {
            alert("Enter username and message!");
            return false;
        } else {
            sendMessage();
        }
    });

    $(window).on("beforeunload", function () {
        console.log("leaving ;(");
        if (socket && socket.readyState === WebSocket.OPEN) {
            const jsonData = { action: "left" };
            socket.send(JSON.stringify(jsonData));
        }
    });

    function updateRecipientList(users) {
        if (!$recipientSelect.length) return;

        const currentUser = $userField.val();
        $recipientSelect.empty();
        $recipientSelect.append(`<option value="">Everyone (Public)</option>`);

        $.each(users, function (index, user) {
            if (user !== currentUser) {
                $recipientSelect.append(`<option value="${user}">${user}</option>`);
            }
        });
    }

    // Display input indicators
    function showTypingIndicator(username) {
        const typingIndicator = `${username} is typing...`;
        $("#typing-indicator").text(typingIndicator).show();
        setTimeout(() => {
            $("#typing-indicator").hide();
        }, 3000);
    }

    //  Clean broadcast message with timestamp
    function displayBroadcastMessage(data) {
        const time = formatTimestamp(data.timestamp);
        const messageHtml = `
            <div class="public-message mb-2 p-2 border rounded">
                <strong>${data.username || "Anonymous"}:</strong> ${data.message}
                <div class="timestamp text-muted"><small>${time}</small></div>
            </div>
        `;
        $output.append(messageHtml);
        $output.scrollTop($output.prop("scrollHeight"));
    }

    //  Private message with timestamp
    function displayPrivateMessage(data) {
        const currentUser = $userField.val();
        const isReceived = data.from !== currentUser;
        const time = formatTimestamp(data.timestamp);

        let messageHtml;
        if (isReceived) {
            messageHtml = `
                <div class="private-message mb-2 p-2 border rounded bg-light">
                    <span class="private-label">Private from <strong>${data.from}</strong>:</span>
                    <span class="message-text"> ${data.message}</span>
                    <div class="timestamp text-muted"><small>${time}</small></div>
                </div>
            `;
        } else {
            messageHtml = `
                <div class="private-message mb-2 p-2 border rounded bg-light">
                    <span class="private-label">🔒 Private to <strong>${data.to}</strong>:</span>
                    <span class="message-text"> ${data.message}</span>
                    <div class="timestamp text-muted"><small>${time}</small></div>
                </div>
            `;
        }

        $output.append(messageHtml);
        $output.scrollTop($output.prop("scrollHeight"));
    }

    //  Error messages stays the same
    function displayError(message) {
        const errorHtml = `
            <div class="error-message">
                <span style="color: #c53030;">❌ ${message}</span>
            </div>
        `;
        $output.append(errorHtml);
        $output.scrollTop($output.prop("scrollHeight"));
    }

    //  Helper to format timestamp to HH:MM
    function formatTimestamp(ts) {
        if (!ts) {
            const now = new Date();
            return `${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}`;
        }
        return ts;
    }

    function sendMessage() {
        const recipient = $recipientSelect.length ? $recipientSelect.val() : "";
        let jsonData;

        if (recipient && recipient !== "") {
            jsonData = {
                action: "private",
                username: $userField.val(),
                to: recipient,
                message: $messageField.val(),
            };
            console.log("Sending private message:", jsonData);
        } else {
            jsonData = {
                action: "broadcast",
                username: $userField.val(),
                message: $messageField.val(),
            };
            console.log("Sending broadcast message:", jsonData);
        }

        socket.send(JSON.stringify(jsonData));
        $messageField.val("");
    }
});
