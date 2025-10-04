let socket = null;
    
$(document).ready(function () {
    const offline = `<span class="badge bg-danger">Not connected</span>`;
    const online = `<span class="badge bg-success">Connected</span>`;

    const $statusDiv = $("#status");
    const $output = $("#output");
    const $userField = $("#username");
    const $messageField = $("#message");
    const $onlineUsers = $("#online_users");
    const $recipientSelect = $("recipient-select");    // recipient selector 

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

                // updating the recipient dropdown 
                updateRecipientList(data.connected_users);
                break;

            case "broadcast":
                // displaying public messages
                $output.append(`<div class="public-message">${data.message}</div>`);
                $output.scrollTop($output.prop("scrollHeight"));
                break;


            case "private":
                 // displaying private messages  
                 displayPrivateMessage(data);
                 break;


            case "error":
                 // displaying error messages 
                 displayError(data.message)
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
   
    // function to update recipient dropdown list 
    function updateRecipientList(users){
         if(!$recipientSelect.length) return;    // skipping if element dosen't exist 

         const currentUser = $userField.val();

         //clear and add default option 
         $recipientSelect.empty();
         $recipientSelect.append(`<option value="">Everyone(Public)</option>`);


         // adding each user except the current user 

         $.each(users , function (index , user) {
             if(user !== currentUser){
                $recipientSelect.append(`<option value="${user}">${user}</option>`); 
             }
         });
    }

    // function to display the private messages 

    function displayPrivateMessage(data){
         const currentUser = $userField.val();
         const isReceived = data.from !== currentUser;


         let messageHtml;
         if(isReceived){
             messageHtml = `
               <div class="private-message">  
                 <span class="private-label">Private from ${data.from}:</span>
                 <span class="message-text">${data.message}</span>
                 </div>
              `;
         } else {
            messageHtml = `
            <div class="private-message">
                <span class="private-label">🔒 Private to ${data.to}:</span>
                <span class="message-text">${data.message}</span>
            </div>
        `;
         }

         $output.append(messageHtml);
         $output.scrollTop($output.prop("scrollHeight"));
    }



    // function to display error messages 

    function displayError(message){
         const errorHtml = `<div class="error-message">
         <span style="color: #c53030;">❌ ${message}</span>
     </div>`;

     $output.append(errorHtml);
     $output.scrollTop($output.prop("scrollHeight"));
    }


    // updating the send message function , now supports the private messaging 
    function sendMessage() {
        const recipient = $recipientSelect.length ? $recipientSelect.val() : "";

        let jsonData;

        if(recipient && recipient !== ""){
             // send private messages 
             jsonData = {
                 action:    "private",
                 username:  $userField.val(),
                 to:        updateRecipientList,
                 message:   $messageField.val(),
             };
             console.log("Sending private message:" , jsonData);
        } else {

            // else send broadcast messages
             jsonData = {
                 action:   "broadcast",
                 username:   $userField.val(),
                 message:   $messageField.val(),
             };
             console.log("Sending broadcast message:" , jsonData);
        }

     
         socket.send(JSON.stringify(jsonData));
         $messageField.val("");
    }
});