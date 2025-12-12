import React, { useState } from "react";

import * as Yup from "yup";
import { useFormik } from "formik";

import { useNavigate } from "react-router";

import type { Databases, PermissionSet } from "../types";


const AskAccess = (): React.ReactElement => {
  const navigate = useNavigate();
  const [step, setStep] = useState(1);

  const [expandedDatabases, setExpandedDatabases] = useState<Record<string, boolean>>({});

  const [databases, setDatabases] = useState<Databases>({});
  const permissionTypes = ["SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER"];

  React.useEffect(() => {
    (async () => {
      try {
        const response = await fetch("/api/databases");
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }
        const data = await response.json();
        setDatabases(data);
      } catch (error) {
        console.error("Error fetching databases:", error);
      }
    })();
  }, []);

  const initialPermissions: PermissionSet = {};
  Object.entries(databases).forEach(([db, tables]) => {
    if (!initialPermissions[db]) {
      initialPermissions[db] = {};
    }

    tables.forEach((table) => {
      initialPermissions[db][table] = {};
      permissionTypes.forEach((perm) => {
        initialPermissions[db][table][perm] = false;
      });
    });
  });

  const validationSchema = Yup.object({
    name: Yup.string().trim().required("Name is required"),
    email: Yup.string().email("Invalid email address").required("Email is required"),
  });

  const formik = useFormik<{
    name: string;
    email: string;
    permissions: PermissionSet;
  }>({
    initialValues: {
      name: "",
      email: "",
      permissions: initialPermissions,
    },
    validationSchema,
    onSubmit: async (values, { setSubmitting }) => {
      try {
        const response = await fetch("/api/request-access", {
          method: "POST",
          body: JSON.stringify(values),
          headers: { "Content-Type": "application/json" },
        });
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }
        const data = await response.json();

        navigate(`/ask-access/${data.id}`);
      } catch (error) {
        console.error("Submission error:", error);
        alert("Failed to submit request. Please try again.");
      } finally {
        setSubmitting(false);
      }
    },
  });

  const handlePermissionChange = (database: string, table: string, permission: string) => {
    const current = formik.values.permissions[database]?.[table]?.[permission] || false;
    formik.setFieldValue(`permissions.${database}.${table}.${permission}`, !current);
  };

  const toggleDatabase = (database: string) => {
    setExpandedDatabases(prev => ({
      ...prev,
      [database]: !prev[database]
    }));
  };

  const handleNext = async (e: React.MouseEvent<HTMLButtonElement>) => {
    e.preventDefault();
    e.stopPropagation();
    await formik.validateForm();
    if (!formik.errors.name && !formik.errors.email && formik.values.name.trim() && formik.values.email.trim()) {
      setStep(step + 1);
    }
  };

  const handlePrev = () => {
    if (step > 1) {
      setStep(step - 1);
    }
  };

  return (
    <div className="flex flex-col justify-center items-center gap-6 p-6 rounded-lg bg-gray-50 border border-gray-200">
      <h2 className="text-2xl font-bold">Request Database Access</h2>

      <div className="w-full max-w-lg">
        <div className="w-full bg-gray-200 rounded-full h-3 relative shadow-inner overflow-hidden">
          <div
            className={`absolute inset-0 h-3 bg-green-500 rounded-full shadow-lg transition-all duration-1000 ease-out transform origin-left ${step === 1 ? "scale-x-50" : "scale-x-100"
              }`}
          />
        </div>
      </div>

      <form onSubmit={formik.handleSubmit} className="w-full">
        {step === 1 && (
          <div className="flex flex-col gap-6 mb-8 w-2xl">
            <div className="flex flex-col w-full">
              <label className="block text-sm font-medium mb-2">Name</label>
              <input
                name="name"
                type="text"
                value={formik.values.name}
                onBlur={formik.handleBlur}
                onChange={formik.handleChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Enter your name"
              />
              {formik.touched.name && formik.errors.name && (
                <p className="mt-1 text-sm text-red-600">{formik.errors.name}</p>
              )}
            </div>
            <div className="flex flex-col w-full">
              <label className="block text-sm font-medium mb-2">Email</label>
              <input
                name="email"
                type="email"
                value={formik.values.email}
                onChange={formik.handleChange}
                onBlur={formik.handleBlur}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Enter your email"
              />
              {formik.touched.email && formik.errors.email && (
                <p className="mt-1 text-sm text-red-600">{formik.errors.email}</p>
              )}
            </div>
          </div>
        )}

        {step === 2 && (
          <div className="mb-8">
            <h3 className="text-lg font-semibold mb-4">Select Required Permissions</h3>
            <div className="overflow-x-auto">
              <table className="w-full border-collapse border border-gray-300">
                <thead>
                  <tr className="bg-gray-100">
                    <th className="border border-gray-300 px-4 py-2 text-left font-semibold">Database / Table</th>
                    {permissionTypes.map((perm) => (
                      <th key={perm} className="border border-gray-300 px-4 py-2 text-center font-semibold">
                        {perm}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {Object.entries(databases).map(([db, tables]) => (
                    <React.Fragment key={db}>
                      <tr className="bg-gray-50 hover:bg-gray-100">
                        <td className="border border-gray-300 px-4 py-2 font-bold">
                          <button
                            type="button"
                            onClick={() => toggleDatabase(db)}
                            className="flex items-center gap-2 text-left w-full hover:text-blue-600"
                          >
                            <span className="text-sm">
                              {expandedDatabases[db] ? '▼' : '▶'}
                            </span>
                            {db}
                          </button>
                        </td>
                        {permissionTypes.map((perm) => (
                          <td key={`${db}-${perm}`} className="border border-gray-300 px-4 py-2 text-center bg-gray-100">
                            {/* No checkboxes at database level */}
                          </td>
                        ))}
                      </tr>
                      {expandedDatabases[db] && tables.map((table) => (
                        <tr key={`${db}-${table}`} className="hover:bg-gray-50">
                          <td className="border border-gray-300 px-8 py-2 text-sm text-gray-700">
                            {table}
                          </td>
                          {permissionTypes.map((perm) => (
                            <td key={`${db}-${table}-${perm}`} className="border border-gray-300 px-4 py-2 text-center">
                              <input
                                type="checkbox"
                                checked={formik.values.permissions[db]?.[table]?.[perm] || false}
                                onChange={() => handlePermissionChange(db, table, perm)}
                                className="w-4 h-4 cursor-pointer"
                              />
                            </td>
                          ))}
                        </tr>
                      ))}
                    </React.Fragment>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* {step === 3 && (
          <div className="flex flex-col items-center gap-6 py-12 text-center">
            <div className="w-20 h-20 bg-green-100 rounded-full flex items-center justify-center shadow-lg">
              <svg className="w-12 h-12 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fillRule="evenodd"
                  d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                  clipRule="evenodd"
                />
              </svg>
            </div>
            <h3 className="text-2xl font-bold text-green-800">Success!</h3>
            <p className="text-lg text-gray-600 max-w-md">Your database access request has been submitted successfully. You will be notified via email once approved.</p>
            <button
              type="button"
              onClick={() => {
                formik.resetForm();
                setStep(1);
              }}
              className="px-8 py-3 bg-green-600 text-white rounded-lg font-semibold hover:bg-green-700 transition-all duration-200 shadow-md"
            >
              Submit New Request
            </button>
          </div>
        )} */}

        {/* Navigation Buttons */}
        {step < 3 && (
          <div className="flex gap-4 justify-between">
            <button
              type="button"
              onClick={handlePrev}
              disabled={step === 1 || formik.isSubmitting}
              className="px-6 py-2 rounded-md font-medium border border-gray-300 hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed transition"
            >
              Previous
            </button>
            {step === 1 ? (
              <button
                type="button"
                onClick={handleNext}
                disabled={!formik.values.name.trim() || !formik.values.email.trim() || formik.isSubmitting}
                className="px-6 py-2 rounded-md font-medium bg-green-600 text-white hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition"
              >
                Next
              </button>
            ) : (
              <button
                type="submit"
                disabled={formik.isSubmitting}
                className="px-6 py-2 rounded-md font-medium bg-green-600 text-white hover:bg-green-700 transition disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {formik.isSubmitting ? "Submitting..." : "Submit Request"}
              </button>
            )}
          </div>
        )}
      </form>
    </div>
  );
};


export default AskAccess;