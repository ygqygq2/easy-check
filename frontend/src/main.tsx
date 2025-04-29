import React from "react";
import { createRoot } from "react-dom/client";
import { createBrowserRouter, Outlet, RouterProvider } from "react-router-dom";

import { Provider } from "@/components/ui/provider";

import { Root } from "./root";
import { routes } from "./routes";

const container = document.getElementById("root");

const root = createRoot(container!);
const router = createBrowserRouter([
  {
    path: "/",
    element: (
      <Root>
        <Outlet />
      </Root>
    ),
    children: [...routes],
  },
]);

root.render(
  <React.StrictMode>
    <Provider>
      <RouterProvider router={router} />
    </Provider>
  </React.StrictMode>
);
