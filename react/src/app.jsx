import React from "react";

function App(){
    const [message, setMessage] = React.useState('');
    const [alerts, setAlerts] = React.useState([]);
    const ws = React.useRef(null);
    const timeout = React.useRef(null);

    const handleChange = e => {
        setMessage(e.target.value);
    }

    const sendMessage = e => {
        e.preventDefault();

        if (ws.current.readyState != 1) {
            return false;
        }
        if (!message) {
            return false;
        }

        let msg = JSON.stringify({
            text: message,
            source: "front",
            time: Date.now()
        });

        ws.current.send(msg);
        addAlert(msg);
        setMessage('');
    }

    const addAlert = (msg) => {
        let item = JSON.parse(msg);
        setAlerts([item, ...alerts.slice(0, 2)]);
    }

    // Sets up a WebSocket connection
    React.useEffect(() => {
        ws.current = new WebSocket("ws://localhost:8080/ws");
        ws.current.onopen = () => {
            console.log('Connected');
            // ws.current.send("Hello from the client!");
        }
        ws.current.onclose = () => console.log('Disconnected');
        
        return () => {
            ws.current.close();
        }

    }, []);

    // Sets a received message handler,
    // only once
    React.useEffect(() => {
        if (!ws.current) return;

        ws.current.onmessage = e => {
            const msg = e.data;
            addAlert(msg);
        }

    }, [alerts]);

    // Clears out alerts with a delay,
    // resets after each message
    React.useEffect(() => {
        clearTimeout(timeout.current);

        timeout.current = setTimeout(() => {setAlerts([])}, 15000);

        return () => {
            clearTimeout(timeout.current);
        }

    }, [alerts]);

    return (
        <div className="container position-relative h-75">
            <div className="position-absolute top-50 start-0 w-100 translate-middle-y">
                <div className="row">
                    <div className="col col-md-8 col-lg-6 mx-auto">
                        <form onSubmit={sendMessage}>
                            <div className="input-group">
                                <input
                                    type="text" className="form-control" placeholder="Enter message" aria-label="Enter message" aria-describedby="button-send"
                                    value={message}
                                    onChange={handleChange}
                                />
                                <button className="btn btn-success" type="submit" id="button-send">Send</button>
                            </div>
                        </form>
                    </div>    
                </div>
            </div>
            <div id="notifications" className="position-absolute top-0 start-0 w-100">
                <div className="row">
                    <div className="col col-md-8 col-lg-6 mx-auto mt-5">
                        {alerts.map((alert) => (
                            <div
                                key={alert.time}
                                className={`alert alert-${alert.source == "back" ? "success" : "primary"}`}
                                role="alert"
                            >
                                <em>{alert.source == "back" ? "Received" : "Sent"}</em>: {alert.text}
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    )
}

export default App;
