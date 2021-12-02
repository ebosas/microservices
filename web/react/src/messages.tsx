import React from "react";
import { Cache, MessageCache } from "./interfaces";

declare global {
  interface Window { __DATA: string | null; }
}
let data: Cache | null = window.__DATA ? JSON.parse(window.__DATA) : null;
window.__DATA = null;

function Messages() {
  const [error, setError] = React.useState(data ? false : null);
  const [isLoaded, setIsLoaded] = React.useState<boolean>(data ? true : false);
  const [messages, setMessages] = React.useState<MessageCache[]>(data ? data.messages : []);
  const [count, setCount] = React.useState<number>(data ? data.count : 0);
  const [total, setTotal] = React.useState<number>(data ? data.total : 0);
  data = null;

  React.useEffect(() => {
    if (isLoaded) return;

    fetch("/api/cache")
      .then(res => res.json())
      .then(
        (result) => {
          setIsLoaded(true);
          setMessages(result.messages);
          setCount(result.count);
          setTotal(result.total);
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
        <h3 className="my-4 ps-2">Recent messages ({count}/{total})</h3>
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
                <td colSpan={3}>No messages</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    );
  }
}

export default Messages;
