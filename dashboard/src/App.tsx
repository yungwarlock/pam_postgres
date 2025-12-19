import React from "react";
import AppRouter from "./router";


const App = (): React.ReactElement => {
  return (
    <div className="flex flex-col">
      <AppRouter />
    </div>
  );
};


export default App
