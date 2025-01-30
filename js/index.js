/** @type {WebSocket} */
var socket;

/** pointer_id: button_number */
const holding_buttons = {};

/** @type {HTMLUListElement} */
var user_list;
/** @type {HTMLSpanElement} */
var user_count_span;
/** @type {HTMLButtonElement[]} */
var buttons;

/**
 * @param {HTMLButtonElement} button 
 * @param {number} number 
 */
function pressed(button, number)
{
    if(button.classList == "") return;
    socket.send(number);
}

/**
 * @param {string} button_status
 */
function get_button_class(button_status)
{
    if(button_status == "0") return "off";
    if(button_status == "1") return "on";
    if(button_status == "-") return "";

    throw new Error("no such character: " + button_status);
}

/**
 * @param {string} data
 */
function get_users(data) {
    const user_text = data.slice(1);
    const users = user_text.split(",");
    users.pop(); // remove last empty entry
    return users
}

/**
 * @param {string} data
 */
function users_change_event(data) {
    const users = get_users(data);
    user_count_span.innerText = users.length;
    user_list.innerHTML = "";
    for(const user of users)
    {
        const li = document.createElement("li");
        li.innerText = user;
        user_list.appendChild(li);
    }
}

/**
 * @param {string} data
 */
function holding_change_event(data) {
    const users = get_users(data);
    const user_button_pairs = [];

    for(const user of users)
    {
        const tmp = user.split(";")
        user_button_pairs[tmp[1]] = tmp[0]
    }

    for(const button of buttons)
    {
        for(const p of button.querySelectorAll("p"))
        {
            button.removeChild(p);
        }

        const user = user_button_pairs[button.getAttribute("pin_num")];
        if(user !== undefined)
        {
            const p = document.createElement("p");
            button.appendChild(p);
            p.innerText = user;
        }
    }
}

/**
 * @param {string} data
 */
function button_change_event(data) {
    for(var i = 0; i < data.length; i++)
    {
        buttons[i].classList = get_button_class(data[i]);
    }
}

function init_buttons() {
    buttons = document.querySelectorAll("#buttons button");
    for(const button of buttons)
    {
        if(button.getAttribute("toggle") != null)
        {
            button.onpointerdown = (e) => {
                if(e.button != 0) return;
    
                const number = button.getAttribute("pin_num");
                pressed(button, number);
            }
            continue;
        }
    
        button.onpointerdown = (e) => {
            if(e.button != 0) return;
    
            if(button.querySelector("p") !== null) return;
            const number = button.getAttribute("pin_num");
            pressed(button, number);
            holding_buttons[e.pointerId] = number;
        }
    }
}

// When page goes out of focus, depress all held button
window.onblur = (e) => {
    for(const k in holding_buttons)
    {
        window.onpointerup({pointerId: k})
    }
}

window.onpointerup = window.onpointercancel = (ev) => {
    const radio_number = holding_buttons[ev.pointerId];
    if(!radio_number) return;

    pressed(document.getElementById(`radio_${radio_number}`), radio_number);
    delete holding_buttons[ev.pointerId];
}

window.onload = () => {
    user_list = document.getElementById("users");
    user_count_span = document.getElementById("user_count");

    init_buttons();

    /** @type {HTMLCanvasElement} */
    const canvas = document.getElementById("video");
    /** @type {CanvasRenderingContext2D} */
    var ctx;
    if(canvas) ctx = canvas.getContext("2d");
    
    var can_recive_frame = true;

    socket = new WebSocket("ws://" + location.host + "/radio_ws");
    socket.binaryType = 'arraybuffer';

    socket.onopen = (event) => {
        console.log("Connected to WebSocket server.");
    };

    socket.onmessage = (event) => {
        const data = event.data;

        if(data instanceof ArrayBuffer) {
            if(!can_recive_frame) return;
            
            can_recive_frame = false;
            const blob = new Blob([data], { type: 'image/jpeg' });
            const img = new Image();
            img.onload = () => {
                if(canvas.hidden) canvas.hidden = false;
                
                ctx.drawImage(img, 0, 0);
                can_recive_frame = true;
                URL.revokeObjectURL(img.src);
            }
            img.onerror = () => {
                console.error("frame dropped");
                can_recive_frame = true;
                URL.revokeObjectURL(img.src);
            };
            img.src = URL.createObjectURL(blob);
            return;
        }

        console.log("Message from server:", data);
        
        if(data === "closed")
        {
            alert("websocket closed")
            return;
        }

        if(data[0] == "u")
        {
            users_change_event(data);
            return;
        }
        else if(data[0] == "h")
        {
            holding_change_event(data);
            return;
        }
        else if(data === "RE")
        {
            alert("Read error");
        }
        else if(data === "WE")
        {
            alert("Write error");
        }

        if(buttons.length !== data.length)
        {
            console.log("wrong length of data");
            return;
        }

        button_change_event(data);
    };

    socket.onclose = (event) => {
        alert("Connection closed. Reloading webpage.");
        window.location.href = window.location.href;
    };
}
