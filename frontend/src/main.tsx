import "./global.css";

import React from "react";
import { createRoot } from "react-dom/client";

import { Provider } from "@/components/ui/provider";

import Root from "./root";

const container = document.getElementById("root");

const root = createRoot(container!);

root.render(
  <React.StrictMode>
    <Provider>
      <Root />
    </Provider>
  </React.StrictMode>,
);
