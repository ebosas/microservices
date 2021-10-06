import React from "react";

function Form({sendMessage}) {
    const [message, setMessage] = React.useState('');

    const submitMessage = (e) => {
        e.preventDefault();

        if (!message) {
            return false;
        }

        if (sendMessage(message)) {
            setMessage('');
        }
    }

    return (
        <form onSubmit={submitMessage}>
            <div className="input-group">
                <input
                    type="text"
                    className="form-control"
                    placeholder="Enter message"
                    aria-label="Enter message"
                    aria-describedby="button-send"
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                />
                <button
                    id="button-send"
                    className="btn btn-success"
                    type="submit"
                >
                    Send
                </button>
            </div>
        </form>
    )
}

export default Form;
