import React from "react";

function Alert({alert}) {
    const alertType = alert.source == "back" ? "success" : "primary";
    const alertLabel = alert.source == "back" ? "Received" : "Sent";

    return (
        <div
            className={"alert alert-"+alertType}
            role="alert"
        >
            <em>{alertLabel}</em>: {alert.text}
        </div>
    );
}

export default Alert;
