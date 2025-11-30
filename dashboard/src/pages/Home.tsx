import React from "react";

import { Link } from "react-router";

const Home = (): React.ReactElement => {
  return (
    <>
      <div className="flex flex-col justify-center items-center px-4">
        <h1 className="text-5xl font-bold mb-4">PAM Postgres</h1>
        <p className="text-xl max-w-4xl text-center">
          Time-constrained and verified elevated privileged access to your Postgres database
        </p>
      </div>

      <div className="grid grid-cols-3 gap-8 px-4 py-12 max-w-5xl mx-auto">
        <div className="flex flex-col justify-center items-center bg-gray-50 border border-gray-200 rounded-lg p-6 h-48">
          <h3 className="text-xl font-semibold mb-2">Principle of Least Privilege</h3>
          <p className="text-slate-500 text-center">Grant only the minimum access necessary for your operations</p>
        </div>
        <div className="flex flex-col justify-center items-center bg-gray-50 border border-gray-200 rounded-lg p-6">
          <h3 className="text-xl font-semibold mb-2">Time-Constrained Access</h3>
          <p className="text-slate-500 text-center">Automatically revoke elevated privileges after a set duration</p>
        </div>
        <div className="flex flex-col justify-center items-center bg-gray-50 border border-gray-200 rounded-lg p-6">
          <h3 className="text-xl font-semibold mb-2">Verified & Secure</h3>
          <p className="text-slate-500 text-center">Reduce vulnerability surface area with secure access controls</p>
        </div>
      </div>

      <div className="flex">
        <Link to="/ask-access" className="transition-colors bg-green-500 hover:bg-green-700 text-white rounded-lg px-6 py-2 text-lg">Ask Access</Link>
      </div>
    </>
  );
};

export default Home;