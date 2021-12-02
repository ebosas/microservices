import React from "react";
import {
  BrowserRouter as Router,
  Switch,
  Route,
  NavLink
} from "react-router-dom";

import Home from './home';
import Messages from './messages';

function App() {
  return (
    <Router>
      <nav className="navbar navbar-expand navbar-light bg-transparent">
        <div className="container-fluid">
          <ul className="navbar-nav">
            <li className="nav-item">
              <NavLink className="nav-link" to="/" activeClassName="active" exact={true}>Home</NavLink>
            </li>
            <li className="nav-item">
              <NavLink className="nav-link" to="/messages" activeClassName="active">Messages</NavLink>
            </li>
          </ul>
        </div>
      </nav>
      <Switch>
        <Route path="/messages">
          <Messages />
        </Route>
        <Route path="/">
          <Home />
        </Route>
      </Switch>
    </Router>
  )
}

export default App;
