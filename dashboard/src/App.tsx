import React from "react";
import AppRouter from "./router";


const App = (): React.ReactElement => {
  return (
    <div className="flex flex-col justify-center items-center w-screen overflow-x-hidden min-h-screen bg-gray-100">
      <AppRouter />
    </div>
  );
};


export default App
