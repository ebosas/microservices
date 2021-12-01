import React from "react";
import Form from './form';
import Alert from './alert';

function Home() {
    const [alerts, setAlerts] = React.useState([]);
    const ws = React.useRef(null);
    const timeout = React.useRef(null);

    const sendMessage = (message) => {
        if (ws.current.readyState != 1) {
            return false;
        }

        let msg = JSON.stringify({
            text: message,
            source: "front",
            time: Date.now()
        });

        ws.current.send(msg);
        addAlert(msg);

        return true;
    }

    const addAlert = (msg) => {
        let item = JSON.parse(msg);
        setAlerts([item, ...alerts.slice(0, 2)]);
    }

    // Sets up a WebSocket connection
    React.useEffect(() => {
        ws.current = new WebSocket("ws://" + window.location.host + "/ws");
        ws.current.onopen = () => {
            console.log('Connected');
            // ws.current.send("Hello from the client!");
        }
        ws.current.onclose = () => console.log('Disconnected');
        ws.current.onerror = () => console.log('Websocket error')
        
        return () => {
            ws.current.close();
        }
    }, []);

    // Sets a received message handler, only once.
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
        timeout.current = setTimeout(() => {setAlerts([])}, 20000);

        return () => {
            clearTimeout(timeout.current);
        }
    }, [alerts]);

    return (
        <div className="container position-relative h-75">
            <div className="position-absolute bottom-50 start-0 w-100 translate-middle-y">
                <div className="row">
                    <div className="col col-md-8 col-lg-6 mx-auto">
                        <Form sendMessage={sendMessage} />
                    </div>
                </div>
            </div>
            <div id="notifications" className="position-absolute top-0 start-0 w-100">
                <div className="row">
                    <div className="col col-md-8 col-lg-6 mx-auto">
                        {alerts.map((alert) => (
                            // TODO: duplicate keys possible
                            <Alert key={alert.time} alert={alert} />
                        ))}
                    </div>
                </div>
            </div>
        </div>
    )
}

export default Home;