import { API } from "aws-amplify";
import { Tenant } from "./Tenant";

function App() {
  async function seed() {
    try {
      const response = await API.get("Endpoint", "/seed");
      console.log(response);
    } catch (error) {
      alert(error.message);
    }
  }
  return (
    <div>
      <Tenant tenantID="1" />
      <hr />
      <Tenant tenantID="2" />
      <hr />
      <button onClick={seed}>Seed data</button>
    </div>
  );
}

export default App;
