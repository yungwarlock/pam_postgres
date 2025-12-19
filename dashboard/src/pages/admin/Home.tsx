import React from "react";
import { Link } from "react-router";


const AdminHome = (): React.ReactElement => {
  // Data for active connections line chart

  const connectionsToday = 20;

  // Data for success rate pie chart
  const todaysRequests = {
    accepted: 180,
    failed: 20
  };

  const recentActivities = [
    {
      title: "User Authentication",
      description: "admin@example.com logged in successfully",
      timestamp: "2 minutes ago",
      color: "blue-500"
    },
    {
      title: "Database Connection",
      description: "New connection established to production database",
      timestamp: "15 minutes ago",
      color: "green-500"
    },
    {
      title: "Permission Update",
      description: "Updated access permissions for user john.doe@example.com",
      timestamp: "1 hour ago",
      color: "yellow-500"
    },
    {
      title: "Schema Change",
      description: "Database schema updated with new tables",
      timestamp: "3 hours ago",
      color: "purple-500"
    },
    {
      title: "Failed Login Attempt",
      description: "Multiple failed login attempts detected from IP 192.168.1.100",
      timestamp: "5 hours ago",
      color: "red-500"
    }
  ];

  return (
    <div className="flex flex-col items-center w-full overflow-x-hidden">
      <div className="flex justify-between items-center w-full border-b border-b-gray-200 py-3 px-8">
        <div className="flex gap-4 items-center">
          <Link to="/">
            <h1 className="text-2xl font-semibold">PAM Postgres</h1>
          </Link>
          <Link to="/admin/connections" className="ml-2">Connections</Link>
          <Link to="/admin/settings" className="">Settings</Link>
        </div>
        <div className="flex gap-4 items-center">
          <div className="w-8 h-8 bg-gray-400 rounded-full"></div>
        </div>
      </div>
      <div className="flex flex-col border-b border-b-gray-200 w-full h-[40vh] px-8 py-6 gap-2">
        <h3 className="text-lg font-semibold mb-4">Today's Overview</h3>
        <div className="flex gap-6 h-full">
          <div className="flex-1 border border-gray-200 rounded-lg p-6 bg-white shadow-sm">
            <h3 className="text-sm font-medium text-gray-500">Active Connections</h3>
            <p className="text-3xl font-bold mt-1">{connectionsToday}</p>
          </div>

          <div className="flex-1 border border-gray-200 rounded-lg p-6 bg-white shadow-sm">
            <h3 className="text-sm font-medium text-gray-500">Total Connections Requests</h3>
            <p className="text-3xl font-bold mt-1">{todaysRequests.accepted} / {todaysRequests.accepted + todaysRequests.failed}</p>
          </div>
        </div>
      </div>

      <div className="flex flex-col justify-center px-8 py-6 overflow-x-hidden w-6xl">
        <h2 className="text-lg font-semibold mb-4">Recent Activity</h2>
        <div className="border border-gray-200 rounded-lg divide-y divide-gray-200 overflow-hidden">
          {recentActivities.map((activity, index) => (
            <div key={index} className={`flex items-start gap-4 p-4 hover:bg-gray-50 transition-colors ${index % 2 === 0 ? 'bg-white' : 'bg-gray-50'}`}>
              <div className={`w-2 h-2 bg-${activity.color} rounded-full mt-2`}></div>
              <div className="flex-1 overflow-x-hidden">
                <div className="flex items-center justify-between">
                  <p className="font-medium text-sm">{activity.title}</p>
                  <span className="text-xs text-gray-500">{activity.timestamp}</span>
                </div>
                <p className="text-sm text-gray-600 mt-1">{activity.description}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};


export default AdminHome;