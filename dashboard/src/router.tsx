import React from "react";
import { Routes, Route } from "react-router";

import Home from "./pages/Home";
import AskAccess from "./pages/AskAccess";
import AdminHome from "./pages/admin/Home";
import WaitForApproval from "./pages/WaitForApproval";
import ListAllAccessRequests from "./pages/ListAllAccessRequests";


const AppRouter = () => {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/admin" element={<AdminHome />} />
      <Route path="/ask-access" element={<AskAccess />} />
      <Route path="/admin/requests" element={<ListAllAccessRequests />} />
      <Route path="/ask-access/:requestID" element={<WaitForApproval />} />
      <Route path="*" element={<div>404 Not Found</div>} />
    </Routes>
  );
};


export default AppRouter;