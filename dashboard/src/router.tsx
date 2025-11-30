import React from "react";
import { Routes, Route } from "react-router";

import Home from "./pages/Home";
import AskAccess from "./pages/AskAccess";
import ListAllAccessRequests from "./pages/ListAllAccessRequests";


const AppRouter = () => {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/ask-access" element={<AskAccess />} />
      <Route path="/admin" element={<ListAllAccessRequests />} />
      <Route path="*" element={<div>404 Not Found</div>} />
    </Routes>
  );
};


export default AppRouter;