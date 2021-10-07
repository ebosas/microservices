import React from "react";

function Messages() {
    const [error, setError] = React.useState(null);
    const [isLoaded, setIsLoaded] = React.useState(false);
    const [items, setItems] = React.useState([]);

    React.useEffect(() => {
        fetch("/api/cache")
          .then(res => res.json())
          .then(
            (result) => {
              setIsLoaded(true);
              setItems(result);
              console.log(result);
            },
            (error) => {
              setIsLoaded(true);
              setError(error);
            }
          )
      }, [])

      return (
        <div className="container">
            <h3 className="my-4 ps-2">Recent messages (3/67)</h3>
            <table className="table">
                <thead>
                    <tr>
                        <th scope="col">Time</th>
                        <th scope="col">Message</th>
                        <th scope="col">Source</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>Mar 3, 2021</td>
                        <td>Hello there from the frontend</td>
                        <td>Front end</td>
                    </tr>
                    <tr>
                        <td>Mar 3, 2021</td>
                        <td>Hello there from the back end</td>
                        <td>Back end</td>
                    </tr>
                    <tr>
                        <td>Mar 3, 2021</td>
                        <td>This is good!</td>
                        <td>Front end</td>
                    </tr>
                </tbody>
            </table>
        </div>
    )
}

export default Messages;
