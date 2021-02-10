import React from "react";
import ReactDOM from "react-dom";
import App from "./App";

import Amplify, { Auth } from "aws-amplify";

Amplify.configure({
  Auth: {
    region: "xxx",
    userPoolId: "xxx",
    userPoolWebClientId: "xxx"
  },
  API: {
    endpoints: [
      {
        name: "Endpoint",
        endpoint: "xxx",
        custom_header: async () => {
          return {
            Authorization: `${(await Auth.currentSession())
              .getAccessToken()
              .getJwtToken()}`
          };
        }
      }
    ]
  }
});

ReactDOM.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
  document.getElementById("root")
);
