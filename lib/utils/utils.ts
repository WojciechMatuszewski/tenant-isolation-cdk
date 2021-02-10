import { path } from "app-root-path";
import { join } from "path";

function getFunctionPath(pathFromFunctionsDir: string) {
  return join(path, `src/dist/functions/${pathFromFunctionsDir}`);
}

export { getFunctionPath };
