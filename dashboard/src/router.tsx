import React from "react";
import { Routes, Route } from "react-router";

import Home from "./pages/Home";
import AskAccess from "./pages/AskAccess";


const AppRouter = () => {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/ask-access" element={<AskAccess />} />
    </Routes>
  );
};


export default AppRouter;