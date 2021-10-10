import React from "react";

let data = window.__DATA || null;
data = data ? JSON.parse(data) : null;
window.__DATA = null;

function Messages() {
    const [error, setError] = React.useState(data ? false : null);
    const [isLoaded, setIsLoaded] = React.useState(data ? true : false);
    const [messages, setMessages] = React.useState(data ? data.messages : []);
    const [counts, setCounts] = React.useState(data ? {count: data.count, total: data.total} : {});

    data = null;

    React.useEffect(() => {
        if (isLoaded) return;

        fetch("/api/cache")
            .then(res => res.json())
            .then(
                (result) => {
                    setIsLoaded(true);
                    setMessages(result.messages);
                    setCounts({count: result.count, total: result.total});
                },
                (error) => {
                    setIsLoaded(true);
                    setError(error);
                }
            )
    }, []);

    if (error) {
        return <div className="container">Something went wrong</div>;
    } else if (!isLoaded) {
        return <div className="container">Loading...</div>;
    } else {
        return (
            <div className="container">
                <h3 className="my-4 ps-2">Recent messages ({counts.count}/{counts.total})</h3>
                <table className="table">
                    <thead>
                        <tr>
                            <th scope="col">Time</th>
                            <th scope="col">Message</th>
                            <th scope="col">Source</th>
                        </tr>
                    </thead>
                    <tbody>
                        {messages.map(msg => (
                            <tr key={msg.time}>
                                <td>{msg.timefmt} ago</td>
                                <td>{msg.text}</td>
                                <td>{msg.source}</td>
                            </tr>
                        ))}
                        {!messages.length && (
                            <tr>
                                <td colSpan="3">No messages</td>
                            </tr>
                        )}
                    </tbody>
                </table>
            </div>
        );
    }
}

export default Messages;
