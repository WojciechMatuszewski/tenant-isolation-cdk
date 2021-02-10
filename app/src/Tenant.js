import React from "react";
import { Auth, API } from "aws-amplify";

export function Tenant({ tenantID }) {
  async function signIn() {
    try {
      const response = await Auth.signIn({
        username: `test${tenantID}@test.pl`,
        password: "test123"
      });

      console.log(response.user);
    } catch (e) {
      alert(`Failed to signin! ${e.message}`);
    }
  }

  async function signUp() {
    try {
      const response = await Auth.signUp({
        username: `test${tenantID}@test.pl`,
        password: "test123",
        attributes: { "custom:tenant": tenantID }
      });

      console.log(response.user);
    } catch (e) {
      alert(`Failed to sign up! ${e.message}`);
    }
  }

  async function makeARequest(tenantID) {
    try {
      const response = await API.get("Endpoint", `/${tenantID}`);
      console.log(response);
    } catch (e) {
      alert(`Failed to  make a request! ${e.message}`);
    }
  }

  function getOtherTenantID() {
    if (tenantID == "1") return "2";
    return "1";
  }

  return (
    <React.Fragment>
      <h1>Tenant {tenantID}</h1>
      <div>
        <button onClick={signUp}>Sign up</button>
        <button onClick={signIn}>Sign in</button>
        <br /> <br />
        <button onClick={() => makeARequest(tenantID)}>
          Current tenant request
        </button>
        <button onClick={() => makeARequest(getOtherTenantID())}>
          Other tenant request
        </button>
      </div>
    </React.Fragment>
  );
}
