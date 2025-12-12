import { useParams } from "react-router";
import React, { useEffect, useState } from "react";

import type { AccessRequest } from "../types";


const WaitForApproval = (): React.ReactElement => {
  const { requestID } = useParams<{ requestID: string }>();
  const [postgresUrls, setPostgresUrls] = useState<string[]>([]);
  const [accessRequest, setAccessRequest] = useState<AccessRequest | null>(null);
  const [status, setStatus] = useState<"waiting" | "approved" | "rejected">("waiting");

  useEffect(() => {
    if (!requestID) return;

    const eventSource = new EventSource(`/api/request-access/${requestID}/sse`);

    eventSource.onmessage = (event) => {
      try {
        const data: AccessRequest = JSON.parse(event.data);
        setAccessRequest(data);
        
        if (data.status === "approved") {
          setStatus("approved");
          const urls = Object.keys(data.permissions).map((dbName) => {
            const auth_details = data.auth_details;
            const sampleUrl = `psql "postgresql://${auth_details.user}:${auth_details.password}@${auth_details.host}:${auth_details.port}/${dbName}"`;
            return sampleUrl;
          });
          setPostgresUrls(urls);
        } else if (data.status === "rejected") {
          setStatus("rejected");
        }
        
        eventSource.close();
      } catch (error) {
        console.error("Error parsing SSE data:", error);
      }
    };

    eventSource.onerror = (error) => {
      console.error("SSE connection error:", error);
      eventSource.close();
    };

    return () => {
      eventSource.close();
    };
  }, [requestID]);

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  return (
    <div className="container mx-auto p-4">
      <div className="max-w-md mx-auto bg-white rounded-lg shadow-md p-6">
        {status === "waiting" && (
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
            <h2 className="text-xl font-semibold mb-2">Waiting for Approval</h2>
            <p className="text-gray-600">Your access request is being reviewed...</p>
          </div>
        )}

        {status === "approved" && (
          <div className="text-center">
            <div className="text-green-500 mb-4">
              <svg className="w-16 h-16 mx-auto" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-green-600 mb-4">Access Approved!</h2>
            <p className="text-gray-600 mb-4">Your request has been approved. Use the PostgreSQL URL below:</p>
            
            <div className="bg-gray-100 p-3 rounded border mb-4">
              {postgresUrls.map((url, index) => (
                <div key={index} className="mb-2">
                  <code className="text-sm break-all">{url}</code>
                  <button
                    onClick={() => copyToClipboard(url)}
                    className="ml-2 bg-blue-500 hover:bg-blue-700 text-white font-bold py-1 px-2 rounded"
                  >
                    Copy URL
                  </button>
                </div>
              ))}
            </div>
            
            {accessRequest && (
              <div className="mt-4 text-left">
                <h3 className="font-semibold mb-2">Access Details:</h3>
                <p><strong>Database:</strong> {JSON.stringify(accessRequest.permissions.database)}</p>
                {/* <p><strong>Tables:</strong> {accessRequest.permissions.tables.join(", ")}</p> */}
                <p><strong>Permissions:</strong> {JSON.stringify(accessRequest.permissions.permissions)}</p>
              </div>
            )}
          </div>
        )}

        {status === "rejected" && (
          <div className="text-center">
            <div className="text-red-500 mb-4">
              <svg className="w-16 h-16 mx-auto" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-red-600 mb-4">Request Rejected</h2>
            <p className="text-gray-600">Your access request has been rejected. Please contact your administrator for more information.</p>
          </div>
        )}
      </div>
    </div>
  );
};

export default WaitForApproval;